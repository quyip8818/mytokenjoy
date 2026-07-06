package usage

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

func CostCNYFromLog(quota int64, modelName string, models []types.ModelInfo) float64 {
	price := newapi.ModelPriceCNY(models, modelName)
	return newapi.CostCNYFromQuota(quota, price)
}

func ResolveConsumeModel(raw store.RawConsumeLog) string {
	return raw.ModelName
}
