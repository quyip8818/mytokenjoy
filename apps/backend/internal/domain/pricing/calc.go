package pricing

import "math"

// CalcQuota computes quota from token counts and unit prices.
// inputPrice/outputPrice: 元/1M tokens
// quotaPerUnit: quota/元
func CalcQuota(inputTokens, outputTokens int64, inputPrice, outputPrice float64, quotaPerUnit int64) int64 {
	costYuan := (float64(inputTokens)*inputPrice + float64(outputTokens)*outputPrice) / 1_000_000
	return int64(math.Ceil(costYuan * float64(quotaPerUnit)))
}
