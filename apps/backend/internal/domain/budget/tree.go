package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) GetTree(ctx context.Context) ([]types.BudgetNode, error) {
	return pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetConsumed(), s.store.Org().Nodes(), s.cfg.Clock())
}

func (s *service) UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error) {
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
		if msg := pkgbudget.ValidateBudgetNodeUpdate(tree, id, budget, reservedValue); msg != nil {
			return domain.Validation(*msg)
		}
		update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
		if !pkgbudget.UpdateBudgetNodeInTree(tree, id, update) {
			return domain.NotFound("Node not found")
		}
		types.ApplyBudgetTreeToOrgNodes(nodes, tree)
		if err := tx.Org().Nodes().SetTree(ctx, nodes); err != nil {
			return fmt.Errorf("persist budget tree: %w", err)
		}
		updated := pkgbudget.FindBudgetNode(tree, id)
		result = *updated
		return nil
	})
	if err == nil {
		s.logger.Info("budget.node.updated", "node_id", id, "budget", budget)
	}
	return result, err
}

func (s *service) ListMemberBudgets(ctx context.Context, deptID string) ([]types.MemberBudgetQuota, error) {
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	if pkgbudget.FindBudgetNode(budgetCtx.Tree, deptID) == nil {
		return nil, domain.NotFound("Department not found")
	}
	quotas := make([]types.MemberBudgetQuota, 0)
	for _, member := range budgetCtx.Members {
		if member.DepartmentID == deptID {
			quotas = append(quotas, pkgbudget.BuildMemberBudgetQuota(member, budgetCtx.PlatformKeys))
		}
	}
	return quotas, nil
}

func (s *service) UpdateMemberBudget(ctx context.Context, memberID string, personalBudget float64) (types.MemberBudgetQuota, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.MemberBudgetQuota{}, err
	}
	var result types.MemberBudgetQuota
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
		s.logger.Info("budget.member.updated", "member_id", memberID, "personal_budget", personalBudget)
	}
	return result, err
}

func (s *service) ApplyAverageBudget(ctx context.Context, deptID string, personalBudget float64, recursive bool) error {
	if personalBudget < 0 {
		return domain.Validation("额度不能为负数")
	}
	return s.store.WithTx(ctx, func(tx store.Store) error {
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

		// Find the target node and collect department IDs to update
		deptIDs := collectDeptIDs(nodes, deptID, recursive)
		if len(deptIDs) == 0 {
			return domain.NotFound("Department not found")
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
		markNodeAvgBudget(nodes, deptID, personalBudget)
		if err := tx.Org().Nodes().SetTree(ctx, nodes); err != nil {
			return fmt.Errorf("persist org tree: %w", err)
		}
		return nil
	})
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

func markNodeAvgBudget(nodes []types.OrgNode, nodeID string, budget float64) {
	var walk func([]types.OrgNode)
	walk = func(list []types.OrgNode) {
		for i := range list {
			if list[i].ID == nodeID {
				list[i].MemberAvgBudget = budget
				return
			}
			if len(list[i].Children) > 0 {
				walk(list[i].Children)
			}
		}
	}
	walk(nodes)
}

func (s *service) GetGroupMemberConsumed(ctx context.Context, groupID string) (map[string]float64, error) {
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return nil, err
	}
	var target *types.BudgetGroup
	for i := range groups {
		if groups[i].ID == groupID {
			target = &groups[i]
			break
		}
	}
	if target == nil {
		return nil, domain.NotFound("Group not found")
	}
	if len(target.MemberIDs) == 0 {
		return make(map[string]float64), nil
	}

	deptID := ""
	if len(target.DepartmentIDs) > 0 {
		deptID = target.DepartmentIDs[0]
	}
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, s.store.Org().Nodes(), deptID, s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	periodKey := open.String()

	// Batch fetch all member consumed values for this period
	allConsumed, err := s.store.BudgetConsumed().ListConsumed(ctx, store.AxisKindMember, periodKey)
	if err != nil {
		return nil, err
	}

	// Filter to only members in this group
	result := make(map[string]float64, len(target.MemberIDs))
	for _, memberID := range target.MemberIDs {
		result[memberID] = allConsumed[memberID] // defaults to 0 if not found
	}
	return result, nil
}
