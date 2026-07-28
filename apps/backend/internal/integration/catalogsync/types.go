package catalogsync

// CatalogVersions holds the remote version of each sync partition.
type CatalogVersions struct {
	Models int `json:"models"`
}

// CatalogModel represents a model entry from the platform catalog.
type CatalogModel struct {
	ModelID      string   `json:"modelId"`
	DisplayName  string   `json:"displayName"`
	Provider     string   `json:"provider"`
	CallType     string   `json:"callType"`
	InputPrice   float64  `json:"inputPrice"`
	OutputPrice  float64  `json:"outputPrice"`
	Capabilities []string `json:"capabilities,omitempty"`
	MaxContext   int      `json:"maxContext,omitempty"`
}

// CatalogResponse is the generic wrapper for a partition catalog response.
type CatalogResponse[T any] struct {
	Version int `json:"version"`
	Data    []T `json:"data"`
}
