package adapter

import (
	"context"

	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

// usageLotConsumer adapts domain/billing/lot functions to the usage.LotConsumer port interface.
type usageLotConsumer struct{}

// NewUsageLotConsumer creates a LotConsumer adapter for the usage domain.
func NewUsageLotConsumer() domainusage.LotConsumer {
	return &usageLotConsumer{}
}

func (a *usageLotConsumer) ConsumeLotsLocked(ctx context.Context, st store.Store, co *store.Company, amount int64) (domainusage.LotConsumeResult, error) {
	result, err := billinglot.ConsumeLotsLocked(ctx, st, co, amount)
	if err != nil {
		return domainusage.LotConsumeResult{}, err
	}
	segs := make([]domainusage.LotSegment, len(result.Segments))
	for i, seg := range result.Segments {
		segs[i] = domainusage.LotSegment{
			LotID:           seg.LotID,
			Quota:           seg.Quota,
			QuotaPerUnit:    seg.QuotaPerUnit,
			DisplayAmount:   seg.DisplayAmount,
			BillingCurrency: seg.BillingCurrency,
		}
	}
	return domainusage.LotConsumeResult{
		Segments:       segs,
		OverdraftUsed:  result.OverdraftUsed,
		OverdraftDelta: result.OverdraftDelta,
	}, nil
}

func (a *usageLotConsumer) LedgerSegmentsFromEntry(base types.UsageLedgerEntry, segs []domainusage.LotSegment) []types.UsageLedgerEntry {
	lotSegs := make([]billinglot.Segment, len(segs))
	for i, seg := range segs {
		lotSegs[i] = billinglot.Segment{
			LotID:           seg.LotID,
			Quota:           seg.Quota,
			QuotaPerUnit:    seg.QuotaPerUnit,
			DisplayAmount:   seg.DisplayAmount,
			BillingCurrency: seg.BillingCurrency,
		}
	}
	return billinglot.LedgerSegmentsFromEntry(base, lotSegs)
}

var _ domainusage.LotConsumer = (*usageLotConsumer)(nil)
