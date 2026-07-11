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

func (s *service) ListApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return nil, err
	}
	return s.store.Budget().BudgetApprovals(ctx)
}

func (s *service) ResolveApproval(ctx context.Context, id string, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return types.BudgetApproval{}, err
	}
	if input.Status != "approved" && input.Status != "rejected" {
		return types.BudgetApproval{}, domain.Validation("invalid status")
	}
	if input.Status == "rejected" && (input.RejectReason == nil || *input.RejectReason == "") {
		return types.BudgetApproval{}, domain.Validation("reject reason required")
	}
	items, err := s.store.Budget().BudgetApprovals(ctx)
	if err != nil {
		return types.BudgetApproval{}, err
	}
	idx := -1
	for i := range items {
		if items[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.BudgetApproval{}, domain.NotFound("Not found")
	}
	if items[idx].Status != "pending" {
		return types.BudgetApproval{}, domain.Validation("approval already resolved")
	}

	approval := items[idx]
	now := time.Now().UTC()

	if input.Status == "approved" {
		deptID := approval.DepartmentID
		if deptID == "" {
			member, err := s.store.Org().MemberByID(ctx, approval.ApplicantID)
			if err != nil {
				return types.BudgetApproval{}, err
			}
			if member == nil {
				return types.BudgetApproval{}, domain.NotFound("申请人不存在")
			}
			deptID = member.DepartmentID
		}

		nodes, err := s.store.Org().Nodes().Tree(ctx)
		if err != nil {
			return types.BudgetApproval{}, err
		}
		tree := types.OrgNodesToBudgetTree(nodes)
		deptNode := pkgbudget.FindBudgetNode(tree, deptID)
		if deptNode == nil {
			return types.BudgetApproval{}, domain.NotFound("部门不存在")
		}

		reserved := float64(0)
		if deptNode.ReservedPool != nil {
			reserved = *deptNode.ReservedPool
		}
		if reserved < approval.Amount {
			return types.BudgetApproval{}, domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %.2f 元", reserved))
		}

		if err := s.store.WithTx(ctx, func(txStore store.Store) error {
			if err := txStore.Budget().AcquireBudgetLock(ctx); err != nil {
				return err
			}
			if err := txStore.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
				return err
			}
			newReserved := reserved - approval.Amount
			deptNode.ReservedPool = &newReserved
			types.ApplyBudgetTreeToOrgNodes(nodes, tree)
			if err := txStore.Org().Nodes().SetTree(ctx, nodes); err != nil {
				return fmt.Errorf("persist budget tree: %w", err)
			}
			members, err := txStore.Org().Members(ctx)
			if err != nil {
				return err
			}
			found := false
			for i := range members {
				if members[i].ID == approval.ApplicantID {
					members[i].PersonalBudget += approval.Amount
					found = true
					break
				}
			}
			if !found {
				return domain.NotFound("申请人不存在")
			}
			if err := txStore.Org().SetMembers(ctx, members); err != nil {
				return fmt.Errorf("persist member personal budget: %w", err)
			}
			return nil
		}); err != nil {
			return types.BudgetApproval{}, err
		}

		if s.enqueueRebalanceAxis != nil {
			_ = s.enqueueRebalanceAxis(ctx, store.RebalanceAxisMember, approval.ApplicantID)
		}
	} else {
		if err := s.store.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
			return types.BudgetApproval{}, err
		}
	}

	resolved := now.Format("2006-01-02 15:04")
	items[idx].Status = input.Status
	items[idx].ResolvedAt = &resolved
	if input.Status == "rejected" {
		items[idx].RejectReason = input.RejectReason
	}
	return items[idx], nil
}
