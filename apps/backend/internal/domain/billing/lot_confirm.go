package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) confirmGiftLot(ctx context.Context, points float64, createdBy string) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := fmt.Sprintf("gift-%d-%d", companyID, now.UnixNano())
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: 0, Currency: currency,
		PointsPerUnit: ppu, PointsGranted: points,
		Source: store.RechargeSourceGift, LotKind: store.LotKindGift,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildGiftLot(order, currency)
	if err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) confirmAdjustLot(ctx context.Context, points, amountDisplay float64, createdBy string) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := fmt.Sprintf("adj-%d-%d", companyID, now.UnixNano())
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: amountDisplay, Currency: currency,
		PointsPerUnit: ppu, PointsGranted: points,
		Source: store.RechargeSourceAdjust, LotKind: store.LotKindAdjust,
		Status: store.RechargeStatusConfirmed, CreatedBy: createdBy,
		CreatedAt: now, UpdatedAt: now,
	}
	lot := BuildAdjustLot(order, currency, amountDisplay)
	if err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) finishPendingOrder(ctx context.Context, order store.RechargeOrder) error {
	co, err := s.store.Company().GetByID(ctx, order.CompanyID)
	if err != nil {
		return err
	}
	if co == nil {
		return domain.NotFound("company not found")
	}
	currency := order.Currency
	if currency == "" {
		currency = resolveBillingCurrency(co)
	}
	ppu := order.PointsPerUnit
	if ppu <= 0 {
		ppu, err = s.lookupPointsPerUnit(ctx, currency)
		if err != nil {
			return err
		}
	}
	if order.PointsGranted <= 0 {
		order.PointsGranted = PointsGrantedFromAmount(order.Amount, ppu)
	}
	order.Currency = currency
	order.LotKind = store.LotKindPaid
	order.Status = store.RechargeStatusConfirmed
	order.PointsPerUnit = ppu
	lot := BuildPaidLot(order, currency, ppu)
	if err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, order.CompanyID)
}

func (s *service) confirmPaidRecharge(ctx context.Context, amount float64, source, createdBy string, idempotencyKey *string) error {
	companyID := company.CompanyID(ctx)
	currency, ppu, err := s.resolveChargeRate(ctx, companyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	orderID := fmt.Sprintf("rch-%d-%d", companyID, now.UnixNano())
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
	if err := billinglot.CreditFromLot(ctx, s.store, order, lot, lot.PointsGranted); err != nil {
		return err
	}
	return s.afterRecharge(ctx, companyID)
}

func (s *service) afterRecharge(ctx context.Context, companyID int64) error {
	if err := s.enqueuer.InsertWalletSync(ctx, companyID); err != nil {
		return err
	}
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil || co.NewAPIWalletUserID == nil {
		return nil
	}
	companyCtx := company.WithContext(ctx, company.Context{
		CompanyID: companyID, NewAPIWalletUserID: *co.NewAPIWalletUserID, Status: co.Status,
	})
	return s.enqueuer.InsertRebalanceCompany(companyCtx, companyID)
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
