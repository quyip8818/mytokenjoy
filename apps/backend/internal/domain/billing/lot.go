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

// BuildPaidLot is a convenience wrapper for paid lots where amountDisplay = order.Amount.
func BuildPaidLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return BuildLot(order, billingCurrency, store.LotKindPaid, order.Amount)
}

func DefaultQuotaPerUnit() int64 {
	return common.DefaultQuotaPerUnit
}
