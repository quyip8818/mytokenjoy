package wallet

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type LotSegment struct {
	LotID           string
	Points          float64
	DisplayAmount   float64
	BillingCurrency string
}

// ConsumeLots is the sole write path for ingest lot consumption + wallet_remain.
func ConsumeLots(ctx context.Context, st store.Store, companyID int64, amountPoint float64) ([]LotSegment, error) {
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
	newRemain := co.WalletRemain - amountPoint + overdraftAdded
	if newRemain < 0 {
		newRemain = 0
	}
	if err := st.Company().SetWalletRemain(ctx, companyID, newRemain, nextHead); err != nil {
		return nil, err
	}
	if remaining > 0 {
		return nil, fmt.Errorf("insufficient lot balance")
	}
	return segments, nil
}

// CreditFromLot is the sole write path for recharge lot insert + wallet_remain delta.
func CreditFromLot(
	ctx context.Context,
	st store.Store,
	order store.RechargeOrder,
	lot store.RechargeLot,
	deltaPoint float64,
) error {
	return st.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Billing().ConfirmRechargeWithLot(ctx, order, lot); err != nil {
			return err
		}
		var fifoHead *string
		if lot.PointsRemaining > 0 && lot.LotKind != store.LotKindOverdraft {
			fifoHead = &lot.ID
		}
		return tx.Company().ApplyWalletDelta(ctx, order.CompanyID, deltaPoint, fifoHead)
	})
}
