package billing

import (
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func PointsGrantedFromAmount(amount float64, pointsPerUnit int64) float64 {
	return amount * float64(pointsPerUnit)
}

func PaidLotDisplayAmount(pointsGranted float64, pointsPerUnit int64) float64 {
	if pointsPerUnit <= 0 {
		return 0
	}
	return pointsGranted / float64(pointsPerUnit)
}

func UnitPriceDisplay(amountDisplay, pointsGranted float64) float64 {
	if pointsGranted <= 0 {
		return 0
	}
	return amountDisplay / pointsGranted
}

func BuildPaidLot(order store.RechargeOrder, billingCurrency string, pointsPerUnit int64) store.RechargeLot {
	amountDisplay := PaidLotDisplayAmount(order.PointsGranted, pointsPerUnit)
	unitPrice := UnitPriceDisplay(amountDisplay, order.PointsGranted)
	return store.RechargeLot{
		ID:               order.ID,
		CompanyID:        order.CompanyID,
		RechargeOrderID:  order.ID,
		BillingCurrency:  billingCurrency,
		LotKind:          store.LotKindPaid,
		AmountDisplay:    amountDisplay,
		PointsGranted:    order.PointsGranted,
		PointsRemaining:  order.PointsGranted,
		UnitPriceDisplay: unitPrice,
		Status:           store.LotStatusActive,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

func BuildAdjustLot(order store.RechargeOrder, billingCurrency string, amountDisplay float64) store.RechargeLot {
	unitPrice := UnitPriceDisplay(amountDisplay, order.PointsGranted)
	return store.RechargeLot{
		ID:               order.ID,
		CompanyID:        order.CompanyID,
		RechargeOrderID:  order.ID,
		BillingCurrency:  billingCurrency,
		LotKind:          store.LotKindAdjust,
		AmountDisplay:    amountDisplay,
		PointsGranted:    order.PointsGranted,
		PointsRemaining:  order.PointsGranted,
		UnitPriceDisplay: unitPrice,
		Status:           store.LotStatusActive,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

func BuildGiftLot(order store.RechargeOrder, billingCurrency string) store.RechargeLot {
	return store.RechargeLot{
		ID:               order.ID,
		CompanyID:        order.CompanyID,
		RechargeOrderID:  order.ID,
		BillingCurrency:  billingCurrency,
		LotKind:          store.LotKindGift,
		AmountDisplay:    0,
		PointsGranted:    order.PointsGranted,
		PointsRemaining:  order.PointsGranted,
		UnitPriceDisplay: 0,
		Status:           store.LotStatusActive,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

func DefaultPointsPerUnit() int64 {
	return common.DefaultPointsPerUnit
}
