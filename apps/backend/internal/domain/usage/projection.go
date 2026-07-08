package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func Apply(ctx context.Context, st store.ConsumptionWriter, entry types.UsageLedgerEntry, snapshotPeriodKey string) error {
	if snapshotPeriodKey == "" {
		return fmt.Errorf("usage projection requires snapshot period key")
	}
	periodKey := snapshotPeriodKey

	if err := st.BudgetSnapshots().IncrementConsumed(ctx, store.SnapshotAxisPlatformKey, entry.PlatformKeyID, periodKey, entry.AmountCNY); err != nil {
		return err
	}
	if entry.BudgetGroupID != nil {
		if err := st.BudgetSnapshots().IncrementConsumed(ctx, store.SnapshotAxisBudgetGroup, *entry.BudgetGroupID, periodKey, entry.AmountCNY); err != nil {
			return err
		}
	}
	if entry.MemberID != nil {
		if err := st.BudgetSnapshots().IncrementConsumed(ctx, store.SnapshotAxisMember, *entry.MemberID, periodKey, entry.AmountCNY); err != nil {
			return err
		}
	}
	if err := st.BudgetSnapshots().RollupOrgNodeAncestors(ctx, entry.DepartmentID, periodKey, entry.AmountCNY); err != nil {
		return err
	}

	var memberID string
	if entry.MemberID != nil {
		memberID = *entry.MemberID
	}
	return st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart:  entry.OccurredAt.UTC().Truncate(time.Hour),
		DepartmentID: entry.DepartmentID,
		MemberID:     memberID,
		Model:        entry.Model,
		CostCNY:      entry.AmountCNY,
		CallCount:    1,
		InputTokens:  entry.InputTokens,
		OutputTokens: entry.OutputTokens,
	})
}
