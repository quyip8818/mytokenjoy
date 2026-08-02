package modelcatalog_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

// TestMergePricingIntoModels verifies the merge logic that ListModelsWithPricing uses:
// models from DB have no price, ratios from NewAPI get converted and merged by modelType.
func TestMergePricingIntoModels(t *testing.T) {
	t.Parallel()

	models := []types.ModelInfo{
		{Type: "gpt-4o", Name: "GPT-4o", Active: true},
		{Type: "claude-3.5-sonnet", Name: "Claude 3.5", Active: true},
		{Type: "no-price-model", Name: "No Price", Active: true},
	}

	ratios := []adminport.ModelPricing{
		{ModelName: "gpt-4o", ModelRatio: 1.25, CompletionRatio: 4.0},
		{ModelName: "claude-3.5-sonnet", ModelRatio: 0.75, CompletionRatio: 5.0},
		// no-price-model has no ratio → should stay 0
	}

	// Build price map (same logic as ListModelsWithPricing)
	priceMap := make(map[string][2]float64, len(ratios))
	for _, r := range ratios {
		input, output := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio)
		priceMap[r.ModelName] = [2]float64{input, output}
	}
	for i := range models {
		if p, ok := priceMap[models[i].Type]; ok {
			models[i].InputPrice = p[0]
			models[i].OutputPrice = p[1]
		}
	}

	// Verify gpt-4o: input=2.5, output=10.0
	if models[0].InputPrice != 2.5 {
		t.Errorf("gpt-4o inputPrice = %f, want 2.5", models[0].InputPrice)
	}
	if models[0].OutputPrice != 10.0 {
		t.Errorf("gpt-4o outputPrice = %f, want 10.0", models[0].OutputPrice)
	}

	// Verify claude: input=1.5, output=7.5
	if models[1].InputPrice != 1.5 {
		t.Errorf("claude inputPrice = %f, want 1.5", models[1].InputPrice)
	}
	if models[1].OutputPrice != 7.5 {
		t.Errorf("claude outputPrice = %f, want 7.5", models[1].OutputPrice)
	}

	// Verify no-price-model stays 0
	if models[2].InputPrice != 0 || models[2].OutputPrice != 0 {
		t.Errorf("no-price-model should have zero prices, got (%f, %f)", models[2].InputPrice, models[2].OutputPrice)
	}
}

// TestMergePricingEmpty verifies merge handles empty ratios gracefully.
func TestMergePricingEmpty(t *testing.T) {
	t.Parallel()

	models := []types.ModelInfo{
		{Type: "gpt-4o", Name: "GPT-4o", Active: true},
	}

	var ratios []adminport.ModelPricing // empty — NewAPI call failed or returned nothing

	priceMap := make(map[string][2]float64, len(ratios))
	for _, r := range ratios {
		input, output := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio)
		priceMap[r.ModelName] = [2]float64{input, output}
	}
	for i := range models {
		if p, ok := priceMap[models[i].Type]; ok {
			models[i].InputPrice = p[0]
			models[i].OutputPrice = p[1]
		}
	}

	// Models should still be returned with zero prices (graceful degradation).
	if models[0].InputPrice != 0 || models[0].OutputPrice != 0 {
		t.Errorf("expected zero prices when no ratios, got (%f, %f)", models[0].InputPrice, models[0].OutputPrice)
	}
}

// TestMergePricingExtraRatios verifies that extra ratios (models not in DB) are harmlessly ignored.
func TestMergePricingExtraRatios(t *testing.T) {
	t.Parallel()

	models := []types.ModelInfo{
		{Type: "gpt-4o", Name: "GPT-4o", Active: true},
	}

	ratios := []adminport.ModelPricing{
		{ModelName: "gpt-4o", ModelRatio: 1.25, CompletionRatio: 4.0},
		{ModelName: "deleted-model", ModelRatio: 99.0, CompletionRatio: 1.0}, // not in DB
	}

	priceMap := make(map[string][2]float64, len(ratios))
	for _, r := range ratios {
		input, output := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio)
		priceMap[r.ModelName] = [2]float64{input, output}
	}
	for i := range models {
		if p, ok := priceMap[models[i].Type]; ok {
			models[i].InputPrice = p[0]
			models[i].OutputPrice = p[1]
		}
	}

	if models[0].InputPrice != 2.5 {
		t.Errorf("gpt-4o inputPrice = %f, want 2.5", models[0].InputPrice)
	}
	if len(models) != 1 {
		t.Errorf("models slice should not grow, got len=%d", len(models))
	}
}
