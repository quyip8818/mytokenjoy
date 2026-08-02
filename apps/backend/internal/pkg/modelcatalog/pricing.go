package modelcatalog

import "github.com/tokenjoy/backend/internal/domain/types"

// PriceFromRatio converts NewAPI model_ratio and completion_ratio into display
// prices (元/1M tokens) used by the backend model catalog.
//
// NewAPI pricing model:
//   - model_ratio represents input cost per 1K tokens in currency units
//   - completion_ratio is the output/input multiplier
//
// Conversion: display_price (元/1M tokens) = ratio * 2
func PriceFromRatio(modelRatio, completionRatio float64) (inputPrice, outputPrice float64) {
	inputPrice = modelRatio * 2
	outputPrice = modelRatio * completionRatio * 2
	return inputPrice, outputPrice
}

// RatioFromPrice converts display prices (元/1M tokens) back to NewAPI ratios.
// Inverse of PriceFromRatio.
func RatioFromPrice(inputPrice, outputPrice float64) (modelRatio, completionRatio float64) {
	if inputPrice == 0 {
		return 0, 0
	}
	modelRatio = inputPrice / 2
	completionRatio = outputPrice / inputPrice
	return modelRatio, completionRatio
}

// MergePricing enriches models with display prices converted from NewAPI ratios.
// Models without a matching ratio entry keep zero price.
// ponytail: shared merge logic used by models service and platform handler.
func MergePricing(models []types.ModelInfo, ratios []RatioEntry) {
	if len(ratios) == 0 {
		return
	}
	priceMap := make(map[string][2]float64, len(ratios))
	for _, r := range ratios {
		input, output := PriceFromRatio(r.ModelRatio, r.CompletionRatio)
		priceMap[r.ModelName] = [2]float64{input, output}
	}
	for i := range models {
		if p, ok := priceMap[models[i].Type]; ok {
			models[i].InputPrice = p[0]
			models[i].OutputPrice = p[1]
		}
	}
}

// RatioEntry is the minimal data needed from NewAPI pricing (avoids importing adminport).
type RatioEntry struct {
	ModelName       string
	ModelRatio      float64
	CompletionRatio float64
}
