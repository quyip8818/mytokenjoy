package catalogsync

// CatalogVersions holds the remote version of each sync partition.
type CatalogVersions struct {
	Models     int `json:"models"`
	Pricing    int `json:"pricing"`
	Discounts  int `json:"discounts"`
	Currencies int `json:"currencies"`
	WalletLots int `json:"walletLots"`
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

// CatalogPricing represents a global pricing entry (no per-company contract).
type CatalogPricing struct {
	ModelType       string  `json:"modelType"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

// CatalogDiscount represents a per-company discount entry from the platform sync API.
type CatalogDiscount struct {
	ModelType string  `json:"modelType"` // exact model type or "*" for wildcard
	Discount  float64 `json:"discount"`  // multiplier: 0.8 = 20% off
}

// CatalogCurrency represents a currency entry from the platform sync API.
type CatalogCurrency struct {
	Code         string `json:"code"`
	QuotaPerUnit int64  `json:"quotaPerUnit"`
}

// CatalogLot represents a lot entry from the platform sync API.
type CatalogLot struct {
	ID              string  `json:"id"`
	OrderID         string  `json:"orderId"`
	LotKind         string  `json:"lotKind"`
	BillingCurrency string  `json:"billingCurrency"`
	QuotaPerUnit    int64   `json:"quotaPerUnit"`
	QuotaGranted    int64   `json:"quotaGranted"`
	QuotaRemaining  int64   `json:"quotaRemaining"`
	PaidAmount      float64 `json:"paidAmount"`
	Status          string  `json:"status"`
	CreatedAt       int64   `json:"createdAt"`
}

// CatalogOrder represents a recharge order entry from the platform sync API.
type CatalogOrder struct {
	ID             string  `json:"id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	QuotaPerUnit   int64   `json:"quotaPerUnit"`
	QuotaGranted   int64   `json:"quotaGranted"`
	Source         string  `json:"source"`
	LotKind        string  `json:"lotKind"`
	Status         string  `json:"status"`
	DisplayOrderID string  `json:"displayOrderId"`
	PaymentMethod  string  `json:"paymentMethod"`
	CreatedAt      int64   `json:"createdAt"`
}

// CatalogLotsResponse is the response from the wallet_lots sync endpoint.
type CatalogLotsResponse struct {
	Data              []CatalogLot   `json:"data"`
	Orders            []CatalogOrder `json:"orders"`
	WalletRemainQuota int64          `json:"walletRemainQuota"`
}
