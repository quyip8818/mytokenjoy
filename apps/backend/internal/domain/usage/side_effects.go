package usage

import (
	"context"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
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
	st store.ConsumptionWriter,
	entry types.UsageLedgerEntry,
	enqueueRebalance func(context.Context, string, string) error,
) error {
	if enqueueRebalance != nil {
		if entry.MemberID != nil {
			if err := enqueueRebalance(ctx, store.RebalanceAxisMember, *entry.MemberID); err != nil {
				return err
			}
		}
		if err := enqueueRebalance(ctx, store.RebalanceAxisDepartment, entry.DepartmentID); err != nil {
			return err
		}
		if entry.BudgetGroupID != nil {
			if err := enqueueRebalance(ctx, store.RebalanceAxisBudgetGroup, *entry.BudgetGroupID); err != nil {
				return err
			}
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
	return st.AsyncJobs().EnqueueOverrun(ctx, overrunRaw)
}
