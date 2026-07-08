package types

const (
	ProviderCustom     = "custom"
	modelCatalogKeySep = "\x1f"
)

func ModelCatalogKey(provider, modelType string) string {
	return provider + modelCatalogKeySep + modelType
}

type ModelInfo struct {
	ModelID      int64    `json:"modelId"`
	CompanyID    int64    `json:"-"`
	Provider     string   `json:"provider"`
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Endpoint     *string  `json:"endpoint,omitempty"`
	InputPrice   float64  `json:"inputPrice"`
	OutputPrice  float64  `json:"outputPrice"`
	MaxContext   int      `json:"maxContext"`
	Enabled      bool     `json:"enabled"`
	Capabilities []string `json:"capabilities"`
}

func (m ModelInfo) IsCustom() bool {
	return m.Provider == ProviderCustom
}

type ModelRef struct {
	ModelID  int64  `json:"modelId"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

type RoutingRule struct {
	ID              string     `json:"id"`
	NodeID          string     `json:"nodeId"`
	NodeName        string     `json:"nodeName"`
	AllowedModelIds []int64    `json:"allowedModelIds"`
	DefaultModelId  *int64     `json:"defaultModelId"`
	FallbackModelId *int64     `json:"fallbackModelId"`
	Inherited       bool       `json:"inherited"`
	AllowedModels   []ModelRef `json:"allowedModels,omitempty"`
	DefaultModel    *ModelRef  `json:"defaultModel,omitempty"`
	FallbackModel   *ModelRef  `json:"fallbackModel,omitempty"`
}

type CreateModelInput struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	BaseURL     string  `json:"baseUrl"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

type ToggleModelInput struct {
	Enabled bool `json:"enabled"`
}

type UpdateModelInput struct {
	Name         *string  `json:"name"`
	Type         *string  `json:"type"`
	Description  *string  `json:"description"`
	Endpoint     *string  `json:"endpoint"`
	InputPrice   *float64 `json:"inputPrice"`
	OutputPrice  *float64 `json:"outputPrice"`
	MaxContext   *int     `json:"maxContext"`
	Capabilities []string `json:"capabilities"`
}

type UpdateRoutingRuleInput struct {
	AllowedModelIds []int64 `json:"allowedModelIds"`
	Inherited       *bool   `json:"inherited"`
	DefaultModelId  *int64  `json:"defaultModelId"`
	FallbackModelId *int64  `json:"fallbackModelId"`
}

type ResolvedWhitelist struct {
	Inherited       bool       `json:"inherited"`
	AllowedModelIds []int64    `json:"allowedModelIds"`
	ParentCount     int        `json:"parentCount"`
	AllowedModels   []ModelRef `json:"allowedModels,omitempty"`
}
