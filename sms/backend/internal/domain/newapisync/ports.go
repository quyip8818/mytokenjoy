package newapisync

import "context"

// AdminPort 定义与 NewAPI 交互的端口
type AdminPort interface {
	// UpsertModelRatio 写入单个模型的定价到 NewAPI
	UpsertModelRatio(ctx context.Context, modelID string, inputPrice, outputPrice float64) error
	// SyncPricing 全量同步多个模型定价（read-modify-write merge）
	SyncPricing(ctx context.Context, entries []PricingEntry) error
	// ListCurrentRatios 读取 NewAPI 当前的 ModelRatio + CompletionRatio
	ListCurrentRatios(ctx context.Context) (map[string]ModelPricing, error)
	// ListModels 读取 NewAPI 的所有模型（从 ModelRatio keys）
	ListModels(ctx context.Context) ([]NewAPIModel, error)
}

// NewAPIModel 表示 NewAPI 中的一个模型
type NewAPIModel struct {
	ModelID         string  `json:"modelId"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	ModelRatio      float64 `json:"modelRatio"`
	CompletionRatio float64 `json:"completionRatio"`
}

// PricingEntry 表示一个模型的定价（元/1M tokens）
type PricingEntry struct {
	ModelID     string  // NewAPI 的 model identifier
	InputPrice  float64 // 元/1M tokens
	OutputPrice float64
}

// ModelPricing 表示 NewAPI 中一个模型的当前定价
type ModelPricing struct {
	ModelID         string  `json:"modelId"`
	ModelRatio      float64 `json:"modelRatio"`
	CompletionRatio float64 `json:"completionRatio"`
	InputPrice      float64 `json:"inputPrice"`  // 换算后: modelRatio * 2
	OutputPrice     float64 `json:"outputPrice"` // 换算后: modelRatio * completionRatio * 2
}
