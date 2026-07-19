package billing

import (
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func BuildPaidLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return store.RechargeLot{
		ID:              order.ID,
		CompanyID:       order.CompanyID,
		RechargeOrderID: order.ID,
		BillingCurrency: billingCurrency,
		LotKind:         store.LotKindPaid,
		AmountDisplay:   order.Amount,
		QuotaPerUnit:    order.QuotaPerUnit,
		QuotaGranted:    order.QuotaGranted,
		QuotaRemaining:  order.QuotaGranted,
		Status:          store.LotStatusActive,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func BuildAdjustLot(order store.RechargeOrder, billingCurrency string, amountDisplay float64) store.RechargeLot {
	return store.RechargeLot{
		ID:              order.ID,
		CompanyID:       order.CompanyID,
		RechargeOrderID: order.ID,
		BillingCurrency: billingCurrency,
		LotKind:         store.LotKindAdjust,
		AmountDisplay:   amountDisplay,
		QuotaPerUnit:    order.QuotaPerUnit,
		QuotaGranted:    order.QuotaGranted,
		QuotaRemaining:  order.QuotaGranted,
		Status:          store.LotStatusActive,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func BuildGiftLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return store.RechargeLot{
		ID:              order.ID,
		CompanyID:       order.CompanyID,
		RechargeOrderID: order.ID,
		BillingCurrency: billingCurrency,
		LotKind:         store.LotKindGift,
		AmountDisplay:   0,
		QuotaPerUnit:    order.QuotaPerUnit,
		QuotaGranted:    order.QuotaGranted,
		QuotaRemaining:  order.QuotaGranted,
		Status:          store.LotStatusActive,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

// BuildMockLot creates a mock lot for simulated funds.
// Mock lots are consumed normally during trial period but expired on upgrade.
func BuildMockLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return store.RechargeLot{
		ID:              order.ID,
		CompanyID:       order.CompanyID,
		RechargeOrderID: order.ID,
		BillingCurrency: billingCurrency,
		LotKind:         store.LotKindMock,
		AmountDisplay:   0,
		QuotaPerUnit:    order.QuotaPerUnit,
		QuotaGranted:    order.QuotaGranted,
		QuotaRemaining:  order.QuotaGranted,
		Status:          store.LotStatusActive,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func DefaultQuotaPerUnit() int64 {
	return common.DefaultQuotaPerUnit
}
