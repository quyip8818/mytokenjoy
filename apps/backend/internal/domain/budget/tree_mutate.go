package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/exchange"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error) {
	if budget < 0 {
		return types.BudgetNode{}, domain.Validation("budget must be non-negative")
	}
	if reservedPool != nil && *reservedPool < 0 {
		return types.BudgetNode{}, domain.Validation("reservedPool must be non-negative")
	}
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetNode{}, err
	}
	var result types.BudgetNode
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		nodes, err := tx.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		tree := types.OrgNodesToBudgetTree(nodes)
		existing := pkgbudget.FindBudgetNode(tree, id)
		if existing == nil {
			return domain.NotFound("Node not found")
		}
		reserved := existing.ReservedPool
		if reservedPool != nil {
			reserved = reservedPool
		}
		reservedValue := 0.0
		if reserved != nil {
			reservedValue = *reserved
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		if msg := pkgbudget.ValidateBudgetNodeUpdate(tree, id, budget, reservedValue, projects, members); msg != nil {
			return domain.Validation(*msg)
		}
		update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
		if !pkgbudget.UpdateBudgetNodeInTree(tree, id, update) {
			return domain.NotFound("Node not found")
		}
		updated := pkgbudget.FindBudgetNode(tree, id)
		if updated == nil {
			return domain.NotFound("Node not found")
		}
		if err := pkgbudget.PersistNodeBudget(ctx, tx.Budget().OrgNodeBudget(), id, *updated); err != nil {
			return fmt.Errorf("persist node budget: %w", err)
		}
		result = *updated
		return nil
	})
	if err == nil {
		s.enqueueCompanyRebalance(ctx, "budget.node")
	}
	return result, err
}

func (s *service) UpdateMemberBudget(ctx context.Context, memberID string, personalBudget float64) (types.MemberBudget, error) {
	if personalBudget < 0 {
		return types.MemberBudget{}, domain.Validation("personalBudget must be non-negative")
	}
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.MemberBudget{}, err
	}
	var result types.MemberBudget
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, tx.BudgetConsumed(), tx.Org(), tx.Budget(), tx.Keys(), s.cfg.Clock())
		if err != nil {
			return err
		}
		if msg := pkgbudget.ValidateMemberBudgetUpdate(budgetCtx.Tree, budgetCtx.Members, budgetCtx.PlatformKeys, memberID, personalBudget); msg != nil {
			return domain.Validation(*msg)
		}
		r, updatedMembers := pkgbudget.ApplyMemberBudgetUpdate(budgetCtx.Members, budgetCtx.PlatformKeys, memberID, personalBudget)
		if err := tx.Org().SetMembers(ctx, updatedMembers); err != nil {
			return fmt.Errorf("persist member personal budget: %w", err)
		}
		result = r
		return nil
	})
	if err == nil {
		s.enqueueMemberRebalance(ctx, memberID, "budget.member")
	}
	return result, err
}

func (s *service) ApplyAverageBudget(ctx context.Context, deptID string, personalBudget float64, recursive bool) error {
	if personalBudget < 0 {
		return domain.Validation("额度不能为负数")
	}
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		nodes, err := tx.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		groups, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}

		// Find root node
		rootNode := findOrgNode(nodes, deptID)
		if rootNode == nil {
			return domain.NotFound("Department not found")
		}

		// Validate root department has budget
		if rootNode.Budget <= 0 {
			return domain.Validation("请先给该部门分配额度")
		}

		// Collect departments to update (skip those with own avg budget set)
		deptIDs := collectDeptIDs(nodes, deptID, recursive)

		// Validate each department has sufficient budget
		var insufficientDepts []string
		for id := range deptIDs {
			node := findOrgNode(nodes, id)
			if node == nil {
				continue
			}
			// For the root dept itself, always validate
			// For child depts in recursive mode, skip if budget is 0
			if id != deptID && node.Budget <= 0 {
				insufficientDepts = append(insufficientDepts, node.Name)
				delete(deptIDs, id)
				continue
			}
			// Calculate: childrenSum + projectSum + memberCount * newAvg
			childrenSum := 0.0
			for _, child := range node.Children {
				childrenSum += child.Budget
			}
			projectSum := pkgbudget.ProjectsBudgetForDept(groups, id)
			memberCount := countMembersInDept(members, id)
			totalAfter := childrenSum + projectSum + float64(memberCount)*personalBudget

			if totalAfter > node.Budget {
				if id == deptID {
					return domain.Validation(fmt.Sprintf(
						"额度不足：设置后成员额度总和（%d人×%s=%s）加上已分配（%s）超出部门总额度（%s）",
						memberCount, exchange.Format(personalBudget), exchange.Format(float64(memberCount)*personalBudget),
						exchange.Format(childrenSum+projectSum), exchange.Format(node.Budget),
					))
				}
				insufficientDepts = append(insufficientDepts, node.Name)
				delete(deptIDs, id)
			}
		}

		// If recursive and some depts are insufficient, report them
		if recursive && len(insufficientDepts) > 0 && len(deptIDs) == 0 {
			return domain.Validation(fmt.Sprintf(
				"以下部门预算不足，请先调整预算后再设置成员额度：%s",
				joinNames(insufficientDepts),
			))
		}

		// Update members in qualifying departments
		updated := false
		for i := range members {
			if !deptIDs[members[i].DepartmentID] {
				continue
			}
			if members[i].PersonalBudget != personalBudget {
				members[i].PersonalBudget = personalBudget
				updated = true
			}
		}
		if updated {
			if err := tx.Org().SetMembers(ctx, members); err != nil {
				return fmt.Errorf("persist member budgets: %w", err)
			}
		}

		// Mark the target department's member_avg_budget
		if err := pkgbudget.PersistMemberAvgBudget(ctx, tx.Budget().OrgNodeBudget(), deptID, personalBudget); err != nil {
			return fmt.Errorf("persist member avg budget: %w", err)
		}

		// If some child depts were skipped due to budget, include a warning in the response
		// (handled via error message that frontend can display)
		return nil
	})
	if err == nil {
		s.enqueueCompanyRebalance(ctx, "budget.avg")
	}
	return err
}

// enqueueCompanyRebalance triggers a full-company rebalance so NewAPI token remain_quota
// reflects the updated budget. Best-effort: failure is logged but does not fail the caller.
func (s *service) enqueueCompanyRebalance(ctx context.Context, source string) {
	companyID := store.CompanyID(ctx)
	if err := s.enqueuer.InsertRebalance(ctx, companyID, store.RebalanceAxisCompany, store.CompanyAxisID(companyID)); err != nil {
		s.logger.Warn(source+".rebalance_enqueue_failed", "error", err)
	}
}

// enqueueMemberRebalance triggers a member-scoped rebalance so the member's token remain_quota
// reflects the updated personal budget.
func (s *service) enqueueMemberRebalance(ctx context.Context, memberID, source string) {
	if err := s.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, memberID); err != nil {
		s.logger.Warn(source+".rebalance_enqueue_failed", "member_id", memberID, "error", err)
	}
}

func findOrgNode(nodes []types.OrgNode, id string) *types.OrgNode {
	var result *types.OrgNode
	var walk func([]types.OrgNode)
	walk = func(list []types.OrgNode) {
		for i := range list {
			if list[i].ID == id {
				result = &list[i]
				return
			}
			if len(list[i].Children) > 0 {
				walk(list[i].Children)
			}
		}
	}
	walk(nodes)
	return result
}

func countMembersInDept(members []types.Member, deptID string) int {
	count := 0
	for _, m := range members {
		if m.DepartmentID == deptID {
			count++
		}
	}
	return count
}

func joinNames(names []string) string {
	if len(names) <= 3 {
		return fmt.Sprintf("%v", names)
	}
	return fmt.Sprintf("%s 等 %d 个部门", names[0], len(names))
}

// collectDeptIDs returns a set of department IDs to apply the budget to.
// If recursive, it includes all descendant departments that haven't set their own avg budget.
func collectDeptIDs(nodes []types.OrgNode, rootID string, recursive bool) map[string]bool {
	result := make(map[string]bool)
	var walk func([]types.OrgNode, bool)
	walk = func(list []types.OrgNode, collecting bool) {
		for _, node := range list {
			if node.ID == rootID {
				result[node.ID] = true
				if recursive && len(node.Children) > 0 {
					walk(node.Children, true)
				}
			} else if collecting {
				// Skip departments that have their own avg budget set
				if node.MemberAvgBudget > 0 {
					continue
				}
				result[node.ID] = true
				if len(node.Children) > 0 {
					walk(node.Children, true)
				}
			} else if len(node.Children) > 0 {
				walk(node.Children, false)
			}
		}
	}
	walk(nodes, false)
	return result
}
