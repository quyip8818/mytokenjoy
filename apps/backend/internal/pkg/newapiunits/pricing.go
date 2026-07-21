package newapiunits

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
