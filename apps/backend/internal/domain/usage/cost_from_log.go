package usage

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func CostFromLog(quota int64, callType string, models []types.ModelInfo, allowedIDs []int64) float64 {
	price := newapiunits.ModelPricePoint(models, allowedIDs, callType)
	return newapiunits.CostFromQuota(quota, price)
}

func ResolveConsumeModel(raw store.RawConsumeLog) string {
	return raw.ModelName
}
