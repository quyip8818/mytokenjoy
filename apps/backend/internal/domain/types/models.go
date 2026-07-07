package types

type ModelInfo struct {
	ID           string   `json:"id"`
	Provider     string   `json:"provider"`
	Name         string   `json:"name"`
	DisplayName  string   `json:"displayName"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Visibility   string   `json:"visibility"`
	Endpoint     *string  `json:"endpoint,omitempty"`
	InputPrice   float64  `json:"inputPrice"`
	OutputPrice  float64  `json:"outputPrice"`
	MaxContext   int      `json:"maxContext"`
	Enabled      bool     `json:"enabled"`
	Capabilities []string `json:"capabilities"`
}

type RoutingRule struct {
	ID            string   `json:"id"`
	NodeID        string   `json:"nodeId"`
	NodeName      string   `json:"nodeName"`
	AllowedModels []string `json:"allowedModels"`
	DefaultModel  *string  `json:"defaultModel"`
	FallbackModel *string  `json:"fallbackModel"`
	Inherited     bool     `json:"inherited"`
}

type CreateModelInput struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	BaseURL     string  `json:"baseUrl"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

type ToggleModelInput struct {
	Enabled bool `json:"enabled"`
}

type UpdateModelInput struct {
	DisplayName  *string  `json:"displayName"`
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Visibility   *string  `json:"visibility"`
	Endpoint     *string  `json:"endpoint"`
	InputPrice   *float64 `json:"inputPrice"`
	OutputPrice  *float64 `json:"outputPrice"`
	MaxContext   *int     `json:"maxContext"`
	Capabilities []string `json:"capabilities"`
}

type UpdateRoutingRuleInput struct {
	AllowedModels []string `json:"allowedModels"`
	Inherited     *bool    `json:"inherited"`
	DefaultModel  *string  `json:"defaultModel"`
	FallbackModel *string  `json:"fallbackModel"`
}

type ResolvedWhitelist struct {
	Inherited     bool     `json:"inherited"`
	AllowedModels []string `json:"allowedModels"`
	ParentCount   int      `json:"parentCount"`
}
