package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/exchange"
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

	parsedID, parseErr := uuid.Parse(id)
	if parseErr != nil {
		return types.BudgetApproval{}, domain.Validation("invalid approval id")
	}

	now := time.Now().UTC()
	var result types.BudgetApproval

	if input.Status == "approved" {
		if err := s.store.WithTx(ctx, func(txStore store.Store) error {
			if err := txStore.Budget().AcquireBudgetLock(ctx); err != nil {
				return err
			}
			// Re-read approvals inside transaction after lock to prevent TOCTOU
			items, err := txStore.Budget().BudgetApprovals(ctx)
			if err != nil {
				return err
			}
			idx := -1
			for i := range items {
				if items[i].ID == parsedID {
					idx = i
					break
				}
			}
			if idx < 0 {
				return domain.NotFound("Not found")
			}
			if items[idx].Status != "pending" {
				return domain.Validation("approval already resolved")
			}
			approval := items[idx]

			deptID := approval.DepartmentID
			if deptID == uuid.Nil {
				member, err := txStore.Org().MemberByID(ctx, approval.ApplicantID)
				if err != nil {
					return err
				}
				if member == nil {
					return domain.NotFound("申请人不存在")
				}
				deptID = member.DepartmentID
			}

			row, found, err := txStore.Budget().OrgNodeBudget().Get(ctx, deptID)
			if err != nil {
				return err
			}
			if !found {
				return domain.NotFound("部门不存在")
			}
			reserved := 0.0
			if row.ReservedPool != nil {
				reserved = *row.ReservedPool
			}
			if reserved < approval.Amount {
				return domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %s", exchange.Format(reserved)))
			}

			if err := txStore.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
				return err
			}
			newReserved := reserved - approval.Amount
			row.ReservedPool = &newReserved
			if err := txStore.Budget().OrgNodeBudget().Upsert(ctx, deptID, row); err != nil {
				return fmt.Errorf("persist reserved pool: %w", err)
			}
			members, err := txStore.Org().Members(ctx)
			if err != nil {
				return err
			}
			memberFound := false
			for i := range members {
				if members[i].ID == approval.ApplicantID {
					members[i].PersonalBudget += approval.Amount
					memberFound = true
					break
				}
			}
			if !memberFound {
				return domain.NotFound("申请人不存在")
			}
			if err := txStore.Org().SetMembers(ctx, members); err != nil {
				return fmt.Errorf("persist member personal budget: %w", err)
			}

			resolved := now.Format("2006-01-02 15:04")
			items[idx].Status = input.Status
			items[idx].ResolvedAt = &resolved
			result = items[idx]
			return nil
		}); err != nil {
			return types.BudgetApproval{}, err
		}

		if err := s.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, result.ApplicantID.String()); err != nil {
			return types.BudgetApproval{}, err
		}
	} else {
		// Rejection path: read approvals to validate, then update status
		items, err := s.store.Budget().BudgetApprovals(ctx)
		if err != nil {
			return types.BudgetApproval{}, err
		}
		idx := -1
		for i := range items {
			if items[i].ID == parsedID {
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
		if err := s.store.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
			return types.BudgetApproval{}, err
		}
		resolved := now.Format("2006-01-02 15:04")
		items[idx].Status = input.Status
		items[idx].ResolvedAt = &resolved
		items[idx].RejectReason = input.RejectReason
		result = items[idx]
	}

	return result, nil
}
