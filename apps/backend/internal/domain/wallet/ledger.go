package wallet

import "github.com/tokenjoy/backend/internal/domain/types"

func LedgerSegmentsFromEntry(base types.UsageLedgerEntry, segs []LotSegment) []types.UsageLedgerEntry {
	out := make([]types.UsageLedgerEntry, 0, len(segs))
	for i, seg := range segs {
		entry := base
		entry.SegmentIndex = i
		entry.LotID = seg.LotID
		entry.Amount = seg.Points
		entry.DisplayAmount = seg.DisplayAmount
		entry.BillingCurrency = seg.BillingCurrency
		if i > 0 {
			entry.CallDetail = types.UsageCallDetail{}
		}
		out = append(out, entry)
	}
	return out
}
