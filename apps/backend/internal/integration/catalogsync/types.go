package catalogsync

// CatalogVersions holds the remote version of each sync partition.
type CatalogVersions struct {
	Models     int `json:"models"`
	Pricing    int `json:"pricing"`
	Currencies int `json:"currencies"`
}

// CatalogModel represents a model entry from the platform catalog.
// ponytail: price fields removed — pricing synced via dedicated /sync/catalog/pricing.
type CatalogModel struct {
	ModelID      string   `json:"modelId"`
	DisplayName  string   `json:"displayName"`
	Provider     string   `json:"provider"`
	CallType     string   `json:"callType"`
	Capabilities []string `json:"capabilities,omitempty"`
	MaxContext   int      `json:"maxContext,omitempty"`
}

// CatalogResponse is the generic wrapper for a partition catalog response.
type CatalogResponse[T any] struct {
	Version int `json:"version"`
	Data    []T `json:"data"`
}

// CatalogPricing represents a pricing entry from the platform sync API.
type CatalogPricing struct {
	ModelType   string  `json:"modelType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
	IsContract  bool    `json:"isContract"`
}

// CatalogCurrency represents a currency entry from the platform sync API.
type CatalogCurrency struct {
	Code         string `json:"code"`
	QuotaPerUnit int64  `json:"quotaPerUnit"`
}
