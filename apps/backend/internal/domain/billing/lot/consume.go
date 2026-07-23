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
	Quota           int64
	QuotaPerUnit    int64   // lot's snapshot rate
	DisplayAmount   float64 // = float64(Quota) / float64(QuotaPerUnit)
	BillingCurrency string
}

// ConsumeResult holds the outcome of lot consumption including overdraft info.
type ConsumeResult struct {
	Segments       []Segment
	OverdraftUsed  bool
	OverdraftDelta int64
}

// ConsumeLots locks the company row and consumes lots. Use when the caller has
// NOT yet acquired the company lock within the current transaction.
func ConsumeLots(ctx context.Context, st CreditStore, companyID uuid.UUID, amount int64) (ConsumeResult, error) {
	co, err := st.Company().LockForUpdate(ctx, companyID)
	if err != nil {
		return ConsumeResult{}, err
	}
	if co == nil {
		return ConsumeResult{}, fmt.Errorf("company not found: %s", companyID)
	}
	return consumeLotsWithCompany(ctx, st, co, amount)
}

// LotStore is the minimal store surface for lot consumption operations.
type LotStore interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
}

// ConsumeLotsLocked consumes lots assuming the company row is already locked
// (i.e. the caller already called Company().LockForUpdate within this tx).
func ConsumeLotsLocked(ctx context.Context, st LotStore, co *store.Company, amount int64) (ConsumeResult, error) {
	if co == nil {
		return ConsumeResult{}, fmt.Errorf("company must not be nil")
	}
	return consumeLotsWithCompany(ctx, st, co, amount)
}

func consumeLotsWithCompany(ctx context.Context, st LotStore, co *store.Company, amount int64) (ConsumeResult, error) {
	companyID := co.ID
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, co.FIFOHeadLotID)
	if err != nil {
		return ConsumeResult{}, err
	}
	remaining := amount
	var segments []Segment
	var nextHead *uuid.UUID
	var overdraftAdded int64
	for _, lotRow := range lots {
		if remaining <= 0 {
			break
		}
		take := lotRow.QuotaRemaining
		if take > remaining {
			take = remaining
		}
		display := common.QuotaToDisplay(take, lotRow.QuotaPerUnit)
		segments = append(segments, Segment{
			LotID:           lotRow.ID,
			Quota:           take,
			QuotaPerUnit:    lotRow.QuotaPerUnit,
			DisplayAmount:   display,
			BillingCurrency: lotRow.BillingCurrency,
		})
		lotRow.QuotaRemaining -= take
		if lotRow.QuotaRemaining <= 0 {
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
			LotID:           od.ID,
			Quota:           remaining,
			QuotaPerUnit:    od.QuotaPerUnit,
			DisplayAmount:   0,
			BillingCurrency: od.BillingCurrency,
		})
		od.QuotaRemaining -= remaining
		if od.QuotaRemaining < 0 {
			od.QuotaRemaining = 0
		}
		if err := st.Billing().UpdateLotRemaining(ctx, *od); err != nil {
			return ConsumeResult{}, err
		}
		remaining = 0
	}
	newRemain := co.WalletQuotaRemain - amount + overdraftAdded
	if newRemain < 0 {
		newRemain = 0
	}
	if err := st.Company().SetWalletQuotaRemain(ctx, companyID, newRemain, nextHead); err != nil {
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

// PostCreditFunc is called after CreditFromLot commits successfully.
// It receives the lot that was just credited. Implementations must not fail
// the overall operation — errors should be logged and swallowed.
type PostCreditFunc func(ctx context.Context, lot store.RechargeLot)

// CreditFromLot is the sole write path for recharge lot insert + wallet_quota_remain delta.
// If onCommit is non-nil it is called after the transaction commits successfully.
func CreditFromLot(
	ctx context.Context,
	st CreditStore,
	order store.RechargeOrder,
	lotRow store.RechargeLot,
	deltaQuota int64,
	onCommit PostCreditFunc,
) error {
	err := st.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Billing().ConfirmRechargeWithLot(ctx, order, lotRow); err != nil {
			return err
		}
		var fifoHead *uuid.UUID
		if lotRow.QuotaRemaining > 0 && lotRow.LotKind != store.LotKindOverdraft {
			co, err := tx.Company().GetByID(ctx, order.CompanyID)
			if err != nil {
				return err
			}
			// Append-only queue: only set head when the queue is empty (first active lot).
			if co.FIFOHeadLotID == nil || *co.FIFOHeadLotID == uuid.Nil {
				fifoHead = &lotRow.ID
			}
		}
		return tx.Company().ApplyWalletDelta(ctx, order.CompanyID, deltaQuota, fifoHead)
	})
	if err != nil {
		return err
	}
	if onCommit != nil {
		onCommit(ctx, lotRow)
	}
	return nil
}
