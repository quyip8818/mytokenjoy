package usage

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

func CostCNYFromLog(quota int64, callType string, models []types.ModelInfo, allowedIDs []int64) float64 {
	price := newapi.ModelPriceCNY(models, allowedIDs, callType)
	return newapi.CostCNYFromQuota(quota, price)
}

func ResolveConsumeModel(raw store.RawConsumeLog) string {
	return raw.ModelName
}
