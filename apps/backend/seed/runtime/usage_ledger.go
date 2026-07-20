package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
)

func ApplyUsageLedger(ctx context.Context, st store.Store, cfg config.Config) error {
	ctx = company.WithContext(ctx, company.Context{CompanyID: contract.DefaultCompanyID})
	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 1})
	if err != nil {
		return fmt.Errorf("count usage ledger: %w", err)
	}
	if len(entries) > 0 {
		return nil
	}
	// Ensure the well-known seed lot exists for usage ledger references.
	if err := ensureSeedLot(ctx, st); err != nil {
		return fmt.Errorf("ensure seed lot: %w", err)
	}
	var snap store.Snapshot
	if cfg.BootstrapIsMinimal() {
		snap = seed.LoadMinimal(cfg)
	} else {
		snap = seed.Load(cfg)
	}
	for _, entry := range snap.UsageLedger {
		if _, err := st.Ledger().InsertOnConflict(ctx, entry); err != nil {
			return fmt.Errorf("seed usage ledger %s: %w", entry.ID, err)
		}
	}
	return nil
}

func ensureSeedLot(ctx context.Context, st store.Store) error {
	companyID := contract.DefaultCompanyID

	// Check if seed lot already exists.
	lot, err := st.Billing().GetLotByID(ctx, contract.IDSeedLot)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if lot != nil {
		return nil
	}

	ppu := common.DefaultQuotaPerUnit
	quota := ppu * 99999
	order := store.RechargeOrder{
		ID:           contract.IDSeedLotOrder,
		CompanyID:    companyID,
		Amount:       0,
		Currency:     common.DefaultBillingCurrency,
		QuotaPerUnit: ppu,
		QuotaGranted: quota,
		Source:       "seed",
		LotKind:      store.LotKindMock,
		Status:       store.RechargeStatusConfirmed,
		CreatedBy:    contract.IDMemberAdmin,
	}
	seedLot := store.RechargeLot{
		ID:              contract.IDSeedLot,
		CompanyID:       companyID,
		RechargeOrderID: contract.IDSeedLotOrder,
		BillingCurrency: common.DefaultBillingCurrency,
		LotKind:         store.LotKindMock,
		AmountDisplay:   0,
		QuotaPerUnit:    ppu,
		QuotaGranted:    quota,
		QuotaRemaining:  quota,
		Status:          store.LotStatusActive,
	}
	if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
		return fmt.Errorf("create seed order: %w", err)
	}
	if err := st.Billing().ConfirmRechargeWithLot(ctx, order, seedLot); err != nil {
		return fmt.Errorf("create seed lot: %w", err)
	}
	return nil
}
