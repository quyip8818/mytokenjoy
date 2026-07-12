package lot

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type Segment struct {
	LotID           string
	Points          float64
	DisplayAmount   float64
	BillingCurrency string
}

// ConsumeLots is the sole write path for ingest lot consumption + wallet_remain.
func ConsumeLots(ctx context.Context, st store.Store, companyID int64, amountPoint float64) ([]Segment, error) {
	co, err := st.Company().LockForUpdate(ctx, companyID)
	if err != nil {
		return nil, err
	}
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, co.FIFOHeadLotID)
	if err != nil {
		return nil, err
	}
	remaining := amountPoint
	var segments []Segment
	var nextHead *string
	overdraftAdded := 0.0
	for _, lotRow := range lots {
		if remaining <= 0 {
			break
		}
		take := lotRow.PointsRemaining
		if take > remaining {
			take = remaining
		}
		display := take * lotRow.UnitPriceDisplay
		segments = append(segments, Segment{
			LotID: lotRow.ID, Points: take, DisplayAmount: display, BillingCurrency: lotRow.BillingCurrency,
		})
		lotRow.PointsRemaining -= take
		if lotRow.PointsRemaining <= 0 {
			lotRow.Status = store.LotStatusExhausted
		} else {
			head := lotRow.ID
			nextHead = &head
		}
		if err := st.Billing().UpdateLotRemaining(ctx, lotRow); err != nil {
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
		segments = append(segments, Segment{
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
	lotRow store.RechargeLot,
	deltaPoint float64,
) error {
	return st.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Billing().ConfirmRechargeWithLot(ctx, order, lotRow); err != nil {
			return err
		}
		var fifoHead *string
		if lotRow.PointsRemaining > 0 && lotRow.LotKind != store.LotKindOverdraft {
			co, err := tx.Company().GetByID(ctx, order.CompanyID)
			if err != nil {
				return err
			}
			// Append-only queue: only set head when the queue is empty (first active lot).
			if co.FIFOHeadLotID == nil || *co.FIFOHeadLotID == "" {
				fifoHead = &lotRow.ID
			}
		}
		return tx.Company().ApplyWalletDelta(ctx, order.CompanyID, deltaPoint, fifoHead)
	})
}
