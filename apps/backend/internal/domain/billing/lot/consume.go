package lot

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Segment struct {
	LotID           uuid.UUID
	Points          float64
	DisplayAmount   float64
	BillingCurrency string
}

// ConsumeResult holds the outcome of lot consumption including overdraft info.
type ConsumeResult struct {
	Segments       []Segment
	OverdraftUsed  bool
	OverdraftDelta float64
}

// ConsumeLots locks the company row and consumes lots. Use when the caller has
// NOT yet acquired the company lock within the current transaction.
func ConsumeLots(ctx context.Context, st CreditStore, companyID uuid.UUID, amountPoint float64) (ConsumeResult, error) {
	co, err := st.Company().LockForUpdate(ctx, companyID)
	if err != nil {
		return ConsumeResult{}, err
	}
	if co == nil {
		return ConsumeResult{}, fmt.Errorf("company not found: %s", companyID)
	}
	return consumeLotsWithCompany(ctx, st, co, amountPoint)
}

// LotStore is the minimal store surface for lot consumption operations.
type LotStore interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
}

// ConsumeLotsLocked consumes lots assuming the company row is already locked
// (i.e. the caller already called Company().LockForUpdate within this tx).
func ConsumeLotsLocked(ctx context.Context, st LotStore, co *store.Company, amountPoint float64) (ConsumeResult, error) {
	if co == nil {
		return ConsumeResult{}, fmt.Errorf("company must not be nil")
	}
	return consumeLotsWithCompany(ctx, st, co, amountPoint)
}

func consumeLotsWithCompany(ctx context.Context, st LotStore, co *store.Company, amountPoint float64) (ConsumeResult, error) {
	companyID := co.ID
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, co.FIFOHeadLotID)
	if err != nil {
		return ConsumeResult{}, err
	}
	remaining := amountPoint
	var segments []Segment
	var nextHead *uuid.UUID
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
			return ConsumeResult{}, err
		}
		remaining -= take
	}
	if remaining > 0 {
		currency := common.ResolveBillingCurrency(co.BillingCurrency)
		overdraftAdded = remaining
		od, err := st.Billing().ExpandOverdraftLot(ctx, companyID, currency, remaining)
		if err != nil {
			return ConsumeResult{}, err
		}
		segments = append(segments, Segment{
			LotID: od.ID, Points: remaining, DisplayAmount: 0, BillingCurrency: od.BillingCurrency,
		})
		od.PointsRemaining -= remaining
		if od.PointsRemaining < 0 {
			od.PointsRemaining = 0
		}
		if err := st.Billing().UpdateLotRemaining(ctx, *od); err != nil {
			return ConsumeResult{}, err
		}
		remaining = 0
	}
	newRemain := co.WalletRemain - amountPoint + overdraftAdded
	if newRemain < 0 {
		newRemain = 0
	}
	if err := st.Company().SetWalletRemain(ctx, companyID, newRemain, nextHead); err != nil {
		return ConsumeResult{}, err
	}
	if remaining > 0 {
		return ConsumeResult{}, fmt.Errorf("insufficient lot balance")
	}
	return ConsumeResult{
		Segments:       segments,
		OverdraftUsed:  overdraftAdded > 0,
		OverdraftDelta: overdraftAdded,
	}, nil
}

// CreditStore is the minimal store surface for lot credit operations.
type CreditStore interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

// CreditFromLot is the sole write path for recharge lot insert + wallet_remain delta.
func CreditFromLot(
	ctx context.Context,
	st CreditStore,
	order store.RechargeOrder,
	lotRow store.RechargeLot,
	deltaPoint float64,
) error {
	return st.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Billing().ConfirmRechargeWithLot(ctx, order, lotRow); err != nil {
			return err
		}
		var fifoHead *uuid.UUID
		if lotRow.PointsRemaining > 0 && lotRow.LotKind != store.LotKindOverdraft {
			co, err := tx.Company().GetByID(ctx, order.CompanyID)
			if err != nil {
				return err
			}
			// Append-only queue: only set head when the queue is empty (first active lot).
			if co.FIFOHeadLotID == nil || *co.FIFOHeadLotID == uuid.Nil {
				fifoHead = &lotRow.ID
			}
		}
		return tx.Company().ApplyWalletDelta(ctx, order.CompanyID, deltaPoint, fifoHead)
	})
}
