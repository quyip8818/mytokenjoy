package usage

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/pricing"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// ApplyTJPricing overrides entry.QuotaAmount with TJ's own pricing calculation.
// Falls back to raw.Quota (no-op) if no pricing row exists for the model.
func ApplyTJPricing(entry types.UsageLedgerEntry, snap EntryBuildSnapshot, tokenJoyCompanyID uuid.UUID) types.UsageLedgerEntry {
	if snap.QuotaPerUnit <= 0 {
		return entry
	}

	p := findPrice(snap.CompanyPricing, entry.Model)
	if p == nil {
		p = findPrice(snap.GlobalPricing, entry.Model)
	}
	if p == nil {
		// ponytail: no pricing configured yet — keep raw.Quota, log warning.
		// Upgrade path: once all models are seeded, this becomes an alert.
		slog.Warn("no model_pricing found, using raw quota", "model", entry.Model)
		return entry
	}

	entry.QuotaAmount = pricing.CalcQuota(
		entry.InputTokens, entry.OutputTokens,
		p.InputPrice, p.OutputPrice, snap.QuotaPerUnit,
	)
	entry.CallDetail.InputPrice = p.InputPrice
	entry.CallDetail.OutputPrice = p.OutputPrice
	entry.CallDetail.ContractPricing = (p.CompanyID != tokenJoyCompanyID)
	return entry
}

func findPrice(prices []store.ModelPricingRow, modelType string) *store.ModelPricingRow {
	for i := range prices {
		if prices[i].ModelType == modelType {
			return &prices[i]
		}
	}
	return nil
}
