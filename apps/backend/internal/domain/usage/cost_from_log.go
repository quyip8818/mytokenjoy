package usage

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func CostCNYFromLog(quota int64, modelName string, models []types.ModelInfo) float64 {
	price := newapi.ModelPriceCNY(models, modelName)
	return newapi.CostCNYFromQuota(quota, price)
}

func ResolveLogModel(modelName string) string {
	if modelName != "" {
		return modelName
	}
	return ""
}

func ResolveWebhookModel(payload newapi.WebhookLogPayload) string {
	if payload.Model != "" {
		return payload.Model
	}
	return newapi.LogEntryModel(newapi.LogEntry{ModelName: payload.Model})
}

func ResolveLogEntryModel(entry newapi.LogEntry) string {
	return newapi.LogEntryModel(entry)
}
