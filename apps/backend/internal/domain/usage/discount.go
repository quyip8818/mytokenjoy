package usage

import (
	"math"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// ApplyDiscount multiplies entry.QuotaAmount by the resolved discount coefficient.
// Passthrough: raw.Quota from NewAPI is trusted; TJ only applies a discount multiplier.
func ApplyDiscount(entry types.UsageLedgerEntry, discounts []store.ModelDiscountRow) types.UsageLedgerEntry {
	d := resolveDiscount(discounts, entry.Model)
	if d == 1.0 {
		return entry
	}
	entry.QuotaAmount = int64(math.Ceil(float64(entry.QuotaAmount) * d))
	entry.CallDetail.Discount = d
	entry.CallDetail.ContractPricing = true
	return entry
}

// resolveDiscount: exact match > wildcard "*" > default 1.0
func resolveDiscount(discounts []store.ModelDiscountRow, modelType string) float64 {
	var wildcard float64 = 1.0
	for _, d := range discounts {
		if d.ModelType == modelType {
			return d.Discount
		}
		if d.ModelType == "*" {
			wildcard = d.Discount
		}
	}
	return wildcard
}
