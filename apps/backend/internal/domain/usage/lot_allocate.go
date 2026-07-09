package usage

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type LotSegment struct {
	LotID           string
	Points          float64
	DisplayAmount   float64
	BillingCurrency string
}

func AllocateConsumptionLots(ctx context.Context, st store.Store, companyID int64, amountPoint float64) ([]LotSegment, error) {
	co, err := st.Company().LockForUpdate(ctx, companyID)
	if err != nil {
		return nil, err
	}
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, co.FIFOHeadLotID)
	if err != nil {
		return nil, err
	}
	remaining := amountPoint
	var segments []LotSegment
	var nextHead *string
	overdraftAdded := 0.0
	for _, lot := range lots {
		if remaining <= 0 {
			break
		}
		take := lot.PointsRemaining
		if take > remaining {
			take = remaining
		}
		display := take * lot.UnitPriceDisplay
		segments = append(segments, LotSegment{
			LotID: lot.ID, Points: take, DisplayAmount: display, BillingCurrency: lot.BillingCurrency,
		})
		lot.PointsRemaining -= take
		if lot.PointsRemaining <= 0 {
			lot.Status = store.LotStatusExhausted
		} else {
			head := lot.ID
			nextHead = &head
		}
		if err := st.Billing().UpdateLotRemaining(ctx, lot); err != nil {
			return nil, err
		}
		remaining -= take
	}
	if remaining > 0 {
		currency := co.BillingCurrency
		if currency == "" {
			currency = "CNY"
		}
		overdraftAdded = remaining
		od, err := st.Billing().ExpandOverdraftLot(ctx, companyID, currency, remaining)
		if err != nil {
			return nil, err
		}
		segments = append(segments, LotSegment{
			LotID: od.ID, Points: remaining, DisplayAmount: 0, BillingCurrency: od.BillingCurrency,
		})
		od.PointsRemaining -= remaining
		if od.PointsRemaining < 0 {
			od.PointsRemaining = 0
		}
		if err := st.Billing().UpdateLotRemaining(ctx, *od); err != nil {
			return nil, err
		}
		remaining = 0
	}
	newBalance := co.BalancePoint - amountPoint + overdraftAdded
	if newBalance < 0 {
		newBalance = 0
	}
	if err := st.Company().UpdateWalletPoint(ctx, companyID, newBalance, nextHead); err != nil {
		return nil, err
	}
	if remaining > 0 {
		return nil, fmt.Errorf("insufficient lot balance")
	}
	return segments, nil
}

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
