package usage

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
)

func CallLogFromLedgerEntry(entry types.UsageLedgerEntry) types.CallLog {
	return types.CallLog{
		ID:             entry.ID,
		Caller:         entry.CallDetail.Caller,
		CallerID:       entry.CallDetail.CallerID,
		CallerType:     entry.CallDetail.CallerType,
		Model:          entry.Model,
		Provider:       entry.CallDetail.Provider,
		InputTokens:    float64(entry.InputTokens),
		OutputTokens:   float64(entry.OutputTokens),
		LatencyMs:      entry.CallDetail.LatencyMs,
		Status:         entry.CallDetail.Status,
		Cost:           entry.Amount,
		CreatedAt:      pkgtime.FormatSyncLog(entry.OccurredAt.UTC()),
		PreviewSnippet: entry.CallDetail.PreviewSnippet,
	}
}

func CallLogsFromLedger(entries []types.UsageLedgerEntry) []types.CallLog {
	items := make([]types.CallLog, len(entries))
	for i, entry := range entries {
		items[i] = CallLogFromLedgerEntry(entry)
	}
	return items
}
