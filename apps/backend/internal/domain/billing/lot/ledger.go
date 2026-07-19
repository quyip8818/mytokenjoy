package lot

import "github.com/tokenjoy/backend/internal/domain/types"

func LedgerSegmentsFromEntry(base types.UsageLedgerEntry, segs []Segment) []types.UsageLedgerEntry {
	out := make([]types.UsageLedgerEntry, 0, len(segs))
	for i, seg := range segs {
		entry := base
		entry.SegmentIndex = i
		entry.LotID = seg.LotID
		entry.Amount = seg.Quota
		// Snapshot at settle time: later quota_per_unit changes must not rewrite these.
		entry.DisplayAmount = seg.DisplayAmount
		entry.BillingCurrency = seg.BillingCurrency
		if i > 0 {
			entry.CallDetail = types.UsageCallDetail{}
		}
		out = append(out, entry)
	}
	return out
}
