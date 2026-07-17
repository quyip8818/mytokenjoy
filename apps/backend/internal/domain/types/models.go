package types

const (
	ProviderCustom     = "custom"
	modelCatalogKeySep = "\x1f"
)

func ModelCatalogKey(provider, modelType string) string {
	return provider + modelCatalogKeySep + modelType
}

type ModelInfo struct {
	ModelID           int64    `json:"modelId"`
	CompanyID         int64    `json:"-"`
	Provider          string   `json:"provider"`
	Type              string   `json:"type"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Endpoint          *string  `json:"endpoint,omitempty"`
	ApiKey            *string  `json:"apiKey,omitempty"`
	EndpointModelName *string  `json:"endpointModelName,omitempty"`
	InputPrice        float64  `json:"inputPrice"`
	OutputPrice       float64  `json:"outputPrice"`
	MaxContext        int      `json:"maxContext"`
	MaxTokens         int      `json:"maxTokens"`
	Enabled           bool     `json:"enabled"`
	Capabilities      []string `json:"capabilities"`
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
	AllowedModelIDs []int64    `json:"allowedModelIds"`
	DefaultModelID  *int64     `json:"defaultModelId"`
	FallbackModelID *int64     `json:"fallbackModelId"`
	Inherited       bool       `json:"inherited"`
	AllowedModels   []ModelRef `json:"allowedModels,omitempty"`
	DefaultModel    *ModelRef  `json:"defaultModel,omitempty"`
	FallbackModel   *ModelRef  `json:"fallbackModel,omitempty"`
}

type CreateModelInput struct {
	Type              string   `json:"type"`
	Name              string   `json:"name"`
	BaseURL           string   `json:"baseUrl"`
	ApiKey            string   `json:"apiKey"`
	EndpointModelName string   `json:"endpointModelName"`
	InputPrice        float64  `json:"inputPrice"`
	OutputPrice       float64  `json:"outputPrice"`
	MaxContext        int      `json:"maxContext"`
	MaxTokens         int      `json:"maxTokens"`
	Capabilities      []string `json:"capabilities"`
}

type ToggleModelInput struct {
	Enabled bool `json:"enabled"`
}

type UpdateModelInput struct {
	Name              *string  `json:"name"`
	Type              *string  `json:"type"`
	Description       *string  `json:"description"`
	Endpoint          *string  `json:"endpoint"`
	ApiKey            *string  `json:"apiKey"`
	EndpointModelName *string  `json:"endpointModelName"`
	InputPrice        *float64 `json:"inputPrice"`
	OutputPrice       *float64 `json:"outputPrice"`
	MaxContext        *int     `json:"maxContext"`
	MaxTokens         *int     `json:"maxTokens"`
	Capabilities      []string `json:"capabilities"`
}

type UpdateRoutingRuleInput struct {
	AllowedModelIDs []int64 `json:"allowedModelIds"`
	Inherited       *bool   `json:"inherited"`
	DefaultModelID  *int64  `json:"defaultModelId"`
	FallbackModelID *int64  `json:"fallbackModelId"`
}

type ResolvedWhitelist struct {
	Inherited       bool       `json:"inherited"`
	AllowedModelIDs []int64    `json:"allowedModelIds"`
	ParentCount     int        `json:"parentCount"`
	AllowedModels   []ModelRef `json:"allowedModels,omitempty"`
}
