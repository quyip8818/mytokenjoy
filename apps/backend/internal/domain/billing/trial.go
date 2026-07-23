package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

// SeedTrialCredit creates a trial lot with simulated funds for a newly registered
// trial company. Call within the registration transaction after company creation.
func SeedTrialCredit(ctx context.Context, st billinglot.CreditStore, companyID uuid.UUID, trialQuota int64) error {
	if trialQuota <= 0 {
		return fmt.Errorf("trial credit amount must be positive")
	}
	currency := common.DefaultBillingCurrency
	ppu := common.DefaultQuotaPerUnit
	now := time.Now().UTC()
	orderID := uuid.Must(uuid.NewV7())

	order := store.RechargeOrder{
		ID:           orderID,
		CompanyID:    companyID,
		Amount:       0,
		Currency:     currency,
		QuotaPerUnit: ppu,
		QuotaGranted: trialQuota,
		Source:       store.RechargeSourceSystem,
		LotKind:      store.LotKindMock,
		Status:       store.RechargeStatusConfirmed,
		CreatedBy:    uuid.Nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	lot := BuildMockLot(order, currency)
	return billinglot.CreditFromLot(ctx, st, order, lot, trialQuota)
}

// TrialUpgradeStore is the minimal store surface needed for the trial→standard upgrade.
type TrialUpgradeStore interface {
	Billing() store.BillingRepository
	Company() store.CompanyRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

// ExpireMockLots expires all active mock lots for a company and recomputes
// wallet_quota_remain based on remaining active (non-mock) lots.
// This is called during the trial→standard upgrade flow.
func ExpireMockLots(ctx context.Context, st TrialUpgradeStore, companyID uuid.UUID) error {
	return st.WithTx(ctx, func(tx store.Store) error {
		// 1. Lock company row to serialize with concurrent ingest/consume.
		co, err := tx.Company().LockForUpdate(ctx, companyID)
		if err != nil {
			return err
		}
		if co == nil {
			return fmt.Errorf("trial upgrade: company %s not found", companyID)
		}

		// 2. Expire all active mock lots.
		_, err = tx.Billing().ExpireMockLots(ctx, companyID)
		if err != nil {
			return err
		}

		// 3. Recompute wallet_quota_remain = sum of remaining active lot quota.
		remain, err := tx.Billing().SumActiveLotsRemaining(ctx, companyID)
		if err != nil {
			return err
		}

		// 4. Update wallet_quota_remain (clear FIFO head if no active lots remain).
		var fifoHead *uuid.UUID
		return tx.Company().SetWalletQuotaRemain(ctx, companyID, remain, fifoHead)
	})
}
