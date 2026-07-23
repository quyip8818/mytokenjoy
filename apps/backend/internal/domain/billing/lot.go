package billing

import (
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

// BuildLot creates a RechargeLot from an order with the given kind and display amount.
func BuildLot(order store.RechargeOrder, billingCurrency string, kind string, amountDisplay float64) store.RechargeLot {
	return store.RechargeLot{
		ID:              order.ID,
		CompanyID:       order.CompanyID,
		RechargeOrderID: order.ID,
		BillingCurrency: billingCurrency,
		LotKind:         kind,
		AmountDisplay:   amountDisplay,
		QuotaPerUnit:    order.QuotaPerUnit,
		QuotaGranted:    order.QuotaGranted,
		QuotaRemaining:  order.QuotaGranted,
		Status:          store.LotStatusActive,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

// Convenience builders — thin wrappers over BuildLot for readability at call sites.

func BuildPaidLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return BuildLot(order, billingCurrency, store.LotKindPaid, order.Amount)
}

func BuildAdjustLot(order store.RechargeOrder, billingCurrency string, amountDisplay float64) store.RechargeLot {
	return BuildLot(order, billingCurrency, store.LotKindAdjust, amountDisplay)
}

func BuildGiftLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return BuildLot(order, billingCurrency, store.LotKindGift, 0)
}

func BuildMockLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return BuildLot(order, billingCurrency, store.LotKindMock, 0)
}

func DefaultQuotaPerUnit() int64 {
	return common.DefaultQuotaPerUnit
}
