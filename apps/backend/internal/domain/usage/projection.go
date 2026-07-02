package usage

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func Apply(ctx context.Context, st store.Store, entry types.UsageLedgerEntry) error {
	if err := st.Keys().AddPlatformKeyUsed(ctx, entry.PlatformKeyID, entry.AmountCNY); err != nil {
		return err
	}
	if entry.BudgetGroupID != nil {
		if err := st.Budget().AddGroupConsumed(ctx, *entry.BudgetGroupID, entry.AmountCNY); err != nil {
			return err
		}
	}
	if err := st.Budget().RollupDepartmentConsumed(ctx, entry.DepartmentID, entry.AmountCNY); err != nil {
		return err
	}
	memberID := ""
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
