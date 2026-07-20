package dashboard

import (
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func bucketFromLedgerEntry(entry types.UsageLedgerEntry) types.UsageBucketRow {
	var memberID uuid.UUID
	if entry.MemberID != nil {
		memberID = *entry.MemberID
	}
	return types.UsageBucketRow{
		BucketStart:   entry.OccurredAt.UTC().Truncate(time.Hour),
		DepartmentID:  entry.DepartmentID,
		MemberID:      memberID,
		Model:         entry.Model,
		QuotaConsumed: entry.Amount,
		DisplayCost:   entry.DisplayAmount,
		CallCount:     1,
		InputTokens:   entry.InputTokens,
		OutputTokens:  entry.OutputTokens,
	}
}

func mergeBucketDelta(dst *types.UsageBucketRow, delta types.UsageBucketRow) {
	if dst.CallCount == 0 {
		*dst = delta
		return
	}
	dst.QuotaConsumed += delta.QuotaConsumed
	dst.DisplayCost += delta.DisplayCost
	dst.CallCount += delta.CallCount
	dst.InputTokens += delta.InputTokens
	dst.OutputTokens += delta.OutputTokens
}
