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
	newRemain := co.WalletRemainQuota - amount + overdraftAdded
	if newRemain < 0 {
		newRemain = 0
	}
	if err := st.Company().SetWalletRemainQuota(ctx, companyID, newRemain, nextHead); err != nil {
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

// PreCreditFunc is called BEFORE the local CreditFromLot transaction commits.
// This is where external side-effects (e.g. syncing quota to NewAPI) should run.
//
// Design rationale ("add first, commit second"):
//   - If PreCreditFunc succeeds but local tx fails → external system has slightly
//     more quota than local. This is safe: the external quota is a loose ceiling,
//     not the source of truth. The extra quota gets consumed naturally or corrected
//     by reconciliation.
//   - If PreCreditFunc fails → local tx never starts, the operation fails cleanly.
//     User sees "recharge failed", retries. No inconsistency.
//   - The reverse (commit first, sync after) is dangerous: local wallet has funds
//     but external system rejects requests due to insufficient quota. User experience
//     is "I paid but can't use the service" — much worse.
type PreCreditFunc func(ctx context.Context, lot store.RechargeLot) error

// CreditFromLot is the sole write path for recharge lot insert + wallet_remain_quota delta.
//
// If a PreCreditFunc is provided, it runs before the transaction. This allows external
// systems (e.g. NewAPI) to be updated first, ensuring the user is never in a state
// where they have local balance but are rejected by the external gateway.
func CreditFromLot(
	ctx context.Context,
	st CreditStore,
	order store.RechargeOrder,
	lotRow store.RechargeLot,
	deltaQuota int64,
	beforeCommit ...PreCreditFunc,
) error {
	// Pre-commit: sync external systems before local state changes.
	if len(beforeCommit) > 0 && beforeCommit[0] != nil {
		if err := beforeCommit[0](ctx, lotRow); err != nil {
			return err
		}
	}

	return st.WithTx(ctx, func(tx store.Store) error {
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
}
