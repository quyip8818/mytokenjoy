package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainwallet "github.com/tokenjoy/backend/internal/domain/wallet"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) confirmGiftLot(ctx context.Context, points float64, createdBy string) error {
	companyID := company.CompanyID(ctx)
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		return fmt.Errorf("company not found")
	}
	now := time.Now().UTC()
	currency := co.BillingCurrency
	if currency == "" {
		currency = "CNY"
	}
	orderID := fmt.Sprintf("gift-%d-%d", companyID, now.UnixNano())
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: 0, Currency: currency,
		PointsPerUnit: DefaultPointsPerUnit(), PointsGranted: points,
		Source: store.RechargeSourceGift, LotKind: store.LotKindGift,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildGiftLot(order, currency)
	if err := domainwallet.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) confirmAdjustLot(ctx context.Context, points, amountDisplay float64, createdBy string) error {
	companyID := company.CompanyID(ctx)
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		return fmt.Errorf("company not found")
	}
	now := time.Now().UTC()
	currency := co.BillingCurrency
	if currency == "" {
		currency = "CNY"
	}
	orderID := fmt.Sprintf("adj-%d-%d", companyID, now.UnixNano())
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amountDisplay, Currency: currency,
		PointsPerUnit: DefaultPointsPerUnit(), PointsGranted: points,
		Source: store.RechargeSourceAdjust, LotKind: store.LotKindAdjust,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildAdjustLot(order, currency, amountDisplay)
	if err := domainwallet.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) finishPendingOrder(ctx context.Context, order store.RechargeOrder) error {
	co, err := s.store.Company().GetByID(ctx, order.CompanyID)
	if err != nil || co == nil {
		return fmt.Errorf("company not found")
	}
	ppu := order.PointsPerUnit
	if ppu <= 0 {
		ppu = DefaultPointsPerUnit()
	}
	if order.PointsGranted <= 0 {
		order.PointsGranted = PointsGrantedFromAmount(order.Amount, ppu)
	}
	order.Currency = co.BillingCurrency
	if order.Currency == "" {
		order.Currency = "CNY"
	}
	order.LotKind = store.LotKindPaid
	order.Status = store.RechargeStatusConfirmed
	order.PointsPerUnit = ppu
	lot := BuildPaidLot(order, co.BillingCurrency, ppu)
	if err := domainwallet.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, order.CompanyID)
}

func (s *service) confirmPaidRecharge(ctx context.Context, amount float64, source, createdBy string, idempotencyKey *string) error {
	companyID := company.CompanyID(ctx)
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		return fmt.Errorf("company not found")
	}
	now := time.Now().UTC()
	ppu := DefaultPointsPerUnit()
	orderID := fmt.Sprintf("rch-%d-%d", companyID, now.UnixNano())
	currency := co.BillingCurrency
	if currency == "" {
		currency = "CNY"
	}
	points := PointsGrantedFromAmount(amount, ppu)
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amount, Currency: currency,
		PointsPerUnit: ppu, PointsGranted: points,
		Source: source, LotKind: store.LotKindPaid,
		IdempotencyKey: idempotencyKey, Status: store.RechargeStatusConfirmed,
		DisplayOrderID: formatDisplayOrderID(now),
		PaymentMethod:  "",
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      createdBy, CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildPaidLot(order, currency, ppu)
	if err := domainwallet.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) afterRecharge(ctx context.Context, companyID int64) error {
	if s.enqueueSync != nil {
		if err := s.enqueueSync(ctx, companyID); err != nil {
			slog.Warn("after recharge: enqueue wallet sync failed", "company_id", companyID, "err", err)
		}
	}
	if s.rebalanceAxis != nil {
		co, err := s.store.Company().GetByID(ctx, companyID)
		if err == nil && co != nil && co.NewAPIWalletUserID != nil {
			companyCtx := company.WithContext(ctx, company.Context{
				CompanyID: companyID, NewAPIWalletUserID: *co.NewAPIWalletUserID, Status: co.Status,
			})
			if err := s.rebalanceAxis(companyCtx, companyID); err != nil {
				slog.Warn("after recharge: enqueue rebalance failed", "company_id", companyID, "err", err)
			}
		}
	}
	return nil
}

func (s *service) ConfirmPayment(ctx context.Context, orderID string) error {
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
