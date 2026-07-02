package budget

import (
	"context"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

type overrunPayload struct {
	DepartmentID   string  `json:"departmentId"`
	MemberID       *string `json:"memberId,omitempty"`
	BudgetGroupID  *string `json:"budgetGroupId,omitempty"`
	PlatformKeyID  string  `json:"platformKeyId"`
}

func enqueueSideEffects(ctx context.Context, st store.Store, entry types.UsageLedgerEntry) error {
	if entry.MemberID != nil {
		if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisMember, *entry.MemberID); err != nil {
			return err
		}
	}
	if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisDepartment, entry.DepartmentID); err != nil {
		return err
	}
	if entry.BudgetGroupID != nil {
		if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisBudgetGroup, *entry.BudgetGroupID); err != nil {
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
	if err := st.Relay().EnqueueOverrun(ctx, overrunRaw); err != nil {
		return err
	}

	if logID, ok := domainusage.ParseNewAPILogID(entry.IdempotencyKey); ok && logID > 0 {
		last, err := st.Relay().GetLastLogID(ctx)
		if err != nil {
			return err
		}
		if logID > last {
			if err := st.Relay().SetLastLogID(ctx, logID); err != nil {
				return err
			}
		}
	}
	return nil
}
