package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) confirmGiftLot(ctx context.Context, amount float64, createdBy uuid.UUID) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := uuid.Must(uuid.NewV7())
	quotaGranted := common.MoneyToQuota(amount, ppu)
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: 0, Currency: currency,
		QuotaPerUnit: ppu, QuotaGranted: quotaGranted,
		Source: store.RechargeSourceGift, LotKind: store.LotKindGift,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildLot(order, currency, store.LotKindGift, 0)
	newRemain, err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.QuotaGranted)
	if err != nil {
		return err
	}
	s.syncWalletBestEffort(ctx, companyID, newRemain)
	return nil
}

func (s *service) confirmAdjustLot(ctx context.Context, amount, paidAmount float64, createdBy uuid.UUID) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := uuid.Must(uuid.NewV7())
	quotaGranted := common.MoneyToQuota(amount, ppu)
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: paidAmount, Currency: currency,
		QuotaPerUnit: ppu, QuotaGranted: quotaGranted,
		Source: store.RechargeSourceAdjust, LotKind: store.LotKindAdjust,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildLot(order, currency, store.LotKindAdjust, paidAmount)
	newRemain, err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.QuotaGranted)
	if err != nil {
		return err
	}
	s.syncWalletBestEffort(ctx, companyID, newRemain)
	return nil
}

func (s *service) finishPendingOrder(ctx context.Context, order store.RechargeOrder) error {
	co, err := s.store.Company().GetByID(ctx, order.CompanyID)
	if err != nil {
		return err
	}
	if co == nil {
		return domain.NotFound("company not found")
	}
	// Prefer order snapshot when present; company currency only fills blanks (order create must have stamped them).
	currency := order.Currency
	if currency == "" {
		currency = resolveBillingCurrency(co)
	}
	ppu := order.QuotaPerUnit
	if ppu <= 0 {
		ppu, err = s.resolveQuotaPerUnit(ctx, currency)
		if err != nil {
			return err
		}
	}
	if order.QuotaGranted <= 0 {
		order.QuotaGranted = common.MoneyToQuota(order.Amount, ppu)
	}
	order.Currency = currency
	order.LotKind = store.LotKindPaid
	order.Status = store.RechargeStatusConfirmed
	order.QuotaPerUnit = ppu
	lot := BuildLot(order, currency, store.LotKindPaid, order.Amount)
	newRemain, err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.QuotaGranted)
	if err != nil {
		return err
	}
	s.syncWalletBestEffort(ctx, order.CompanyID, newRemain)
	return nil
}

func (s *service) confirmPaidRecharge(ctx context.Context, amount float64, source string, createdBy uuid.UUID, idempotencyKey *string) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := uuid.Must(uuid.NewV7())
	quotaGranted := common.MoneyToQuota(amount, ppu)
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Currency: currency,
		QuotaPerUnit: ppu, QuotaGranted: quotaGranted,
		Source: source, LotKind: store.LotKindPaid,
		IdempotencyKey: idempotencyKey, Status: store.RechargeStatusConfirmed,
		DisplayOrderID: formatDisplayOrderID(now),
		PaymentMethod:  "",
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      createdBy, CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildLot(order, currency, store.LotKindPaid, order.Amount)
	newRemain, err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.QuotaGranted)
	if err != nil {
		return err
	}
	s.syncWalletBestEffort(ctx, companyID, newRemain)
	return nil
}

func (s *service) ConfirmPayment(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.store.Billing().GetRechargeOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return domain.NotFound("order not found")
	}
	if order.CompanyID != company.CompanyID(ctx) {
		return domain.Forbidden("order does not belong to current company")
	}
	if order.Status == store.RechargeStatusConfirmed {
		return nil
	}
	return s.finishPendingOrder(ctx, *order)
}

// syncWalletBestEffort overrides the NewAPI wallet to match the local SOT (best-effort, post-commit).
func (s *service) syncWalletBestEffort(ctx context.Context, companyID uuid.UUID, newRemain int64) {
	// Bump wallet_lots catalog version so Local sync picks up changes.
	_, _ = s.store.SystemSettings().Increment(ctx, "catalog.wallet_lots_version")

	if s.quotaSyncer == nil {
		return
	}
	walletUserID, ok := company.ResolveNewAPIWalletCompanyID(ctx, s.store.Company())
	if !ok {
		return
	}
	if err := s.quotaSyncer.ManageUser(ctx, walletUserID, "set_quota", newRemain); err != nil {
		slog.Default().Warn("wallet sync after recharge failed",
			"company_id", companyID, "new_remain", newRemain, "error", err)
	}
}
