package usage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type overrunPayload struct {
	DepartmentID  string  `json:"departmentId"`
	MemberID      *string `json:"memberId,omitempty"`
	BudgetGroupID *string `json:"budgetGroupId,omitempty"`
	PlatformKeyID string  `json:"platformKeyId"`
}

func enqueueSideEffects(
	ctx context.Context,
	tx store.Tx,
	entry types.UsageLedgerEntry,
	enqueuer jobs.Enqueuer,
) error {
	if enqueuer == nil {
		return nil
	}
	if tx == nil {
		return fmt.Errorf("enqueue side effects: transaction required")
	}
	companyID := company.CompanyID(ctx)

	if entry.MemberID != nil {
		if err := jobs.InsertRebalance(ctx, enqueuer, tx, companyID, store.RebalanceAxisMember, *entry.MemberID); err != nil {
			return err
		}
	}
	if err := jobs.InsertRebalance(ctx, enqueuer, tx, companyID, store.RebalanceAxisDepartment, entry.DepartmentID); err != nil {
		return err
	}
	if entry.BudgetGroupID != nil {
		if err := jobs.InsertRebalance(ctx, enqueuer, tx, companyID, store.RebalanceAxisBudgetGroup, *entry.BudgetGroupID); err != nil {
			return err
		}
	}

	overrunRaw, err := json.Marshal(overrunPayload{
		DepartmentID:  entry.DepartmentID,
		MemberID:      entry.MemberID,
		BudgetGroupID: entry.BudgetGroupID,
		PlatformKeyID: entry.PlatformKeyID,
	})
	if err != nil {
		return err
	}
	return jobs.InsertOverrun(ctx, enqueuer, tx, companyID, overrunRaw)
}
