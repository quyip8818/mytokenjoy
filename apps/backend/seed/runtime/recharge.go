package runtime

import (
	"context"
	"fmt"
	"time"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func ApplyRechargeOrders(ctx context.Context, st store.Store) error {
	if _, ok := company.FromContext(ctx); !ok {
		ctx = company.DefaultContext(contract.DefaultCompanyID)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, contract.DefaultCompanyID)
	if err != nil {
		return fmt.Errorf("list recharge orders: %w", err)
	}
	if len(orders) > 0 {
		return nil
	}
	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil {
		return fmt.Errorf("load company for seed recharge: %w", err)
	}
	if co == nil {
		return fmt.Errorf("company %d not found for seed recharge", contract.DefaultCompanyID)
	}
	currency := common.ResolveBillingCurrency(co.BillingCurrency)
	cur, err := st.Billing().GetCurrency(ctx, currency)
	if err != nil {
		return fmt.Errorf("load currency %s: %w", currency, err)
	}
	ppu := domainbilling.DefaultPointsPerUnit()
	if cur != nil && cur.PointsPerUnit > 0 {
		ppu = cur.PointsPerUnit
	}
	for _, order := range buildSeedRechargeOrders() {
		order.Currency = currency
		order.PointsPerUnit = ppu
		order.PointsGranted = domainbilling.PointsGrantedFromAmount(order.Amount, ppu)
		order.LotKind = store.LotKindPaid
		if order.Status == store.RechargeStatusConfirmed {
			lot := domainbilling.BuildPaidLot(order, currency, ppu)
			if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
				return fmt.Errorf("seed recharge lot %s: %w", order.ID, err)
			}
			continue
		}
		if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
			return fmt.Errorf("seed recharge order %s: %w", order.ID, err)
		}
	}
	return nil
}

func buildSeedRechargeOrders() []store.RechargeOrder {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	specs := []struct {
		id             string
		amount         float64
		status         string
		createdAt      string
		displayOrderID string
		paymentMethod  string
		invoiceStatus  string
	}{
		{"tu-1", 100, store.RechargeStatusConfirmed, "2026-06-19 14:30:00", "ORD202606190001", store.PaymentMethodAlipay, store.InvoiceStatusNone},
		{"tu-2", 50, store.RechargeStatusConfirmed, "2026-06-18 10:15:00", "ORD202606180002", store.PaymentMethodWechat, store.InvoiceStatusApplied},
		{"tu-3", 200, store.RechargeStatusConfirmed, "2026-06-15 09:00:00", "ORD202606150003", store.PaymentMethodAlipay, store.InvoiceStatusIssued},
		{"tu-4", 20, store.RechargeStatusPending, "2026-06-12 16:45:00", "ORD202606120004", store.PaymentMethodWechat, store.InvoiceStatusNone},
		{"tu-5", 500, store.RechargeStatusConfirmed, "2026-06-10 08:20:00", "ORD202606100005", store.PaymentMethodAlipay, store.InvoiceStatusIssued},
	}
	orders := make([]store.RechargeOrder, 0, len(specs))
	for _, spec := range specs {
		createdAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", spec.createdAt, loc)
		if parseErr != nil {
			createdAt = time.Now().UTC()
		}
		order := store.RechargeOrder{
			ID:             spec.id,
			CompanyID:      contract.DefaultCompanyID,
			Amount:         spec.amount,
			Source:         store.RechargeSourceSelf,
			Status:         spec.status,
			DisplayOrderID: spec.displayOrderID,
			PaymentMethod:  spec.paymentMethod,
			InvoiceStatus:  spec.invoiceStatus,
			CreatedBy:      contract.IDMemberAdmin,
			CreatedAt:      createdAt.UTC(),
			UpdatedAt:      createdAt.UTC(),
		}
		orders = append(orders, order)
	}
	return orders
}
