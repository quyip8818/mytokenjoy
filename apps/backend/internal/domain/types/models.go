package types

import "github.com/google/uuid"

const (
	ProviderCustom     = "custom"
	modelCatalogKeySep = "\x1f"
)

func ModelCatalogKey(provider, modelType string) string {
	return provider + modelCatalogKeySep + modelType
}

type ModelInfo struct {
	ID                uuid.UUID `json:"modelId"`
	CompanyID         uuid.UUID `json:"-"`
	Provider          string    `json:"provider"`
	Type              string    `json:"type"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Endpoint          *string   `json:"endpoint,omitempty"`
	ApiKey            *string   `json:"apiKey,omitempty"`
	EndpointModelName *string   `json:"endpointModelName,omitempty"`
	InputPrice        float64   `json:"inputPrice"`
	OutputPrice       float64   `json:"outputPrice"`
	MaxContext        int       `json:"maxContext"`
	MaxTokens         int       `json:"maxTokens"`
	Enabled           bool      `json:"enabled"`
	Capabilities      []string  `json:"capabilities"`
}

func (m ModelInfo) IsCustom() bool {
	return m.Provider == ProviderCustom
}

type ModelRef struct {
	ID       uuid.UUID `json:"modelId"`
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Provider string    `json:"provider"`
	Enabled  bool      `json:"enabled"`
}

type RoutingRule struct {
	ID              uuid.UUID   `json:"id"`
	NodeID          uuid.UUID   `json:"nodeId"`
	NodeName        string      `json:"nodeName"`
	AllowedModelIDs []uuid.UUID `json:"allowedModelIds"`
	DefaultModelID  *uuid.UUID  `json:"defaultModelId"`
	FallbackModelID *uuid.UUID  `json:"fallbackModelId"`
	Inherited       bool        `json:"inherited"`
	AllowedModels   []ModelRef  `json:"allowedModels,omitempty"`
	DefaultModel    *ModelRef   `json:"defaultModel,omitempty"`
	FallbackModel   *ModelRef   `json:"fallbackModel,omitempty"`
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
	AllowedModelIDs []uuid.UUID `json:"allowedModelIds"`
	Inherited       *bool       `json:"inherited"`
	DefaultModelID  *uuid.UUID  `json:"defaultModelId"`
	FallbackModelID *uuid.UUID  `json:"fallbackModelId"`
}

type ResolvedWhitelist struct {
	Inherited       bool        `json:"inherited"`
	AllowedModelIDs []uuid.UUID `json:"allowedModelIds"`
	ParentCount     int         `json:"parentCount"`
	AllowedModels   []ModelRef  `json:"allowedModels,omitempty"`
}
