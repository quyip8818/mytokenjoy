package sms

// PartitionVersions holds the remote version of each sync partition.
type PartitionVersions struct {
	Channels   int `json:"channels"`
	Models     int `json:"models"`
	Currencies int `json:"currencies"`
}

// CatalogChannel represents a channel entry from the SMS catalog.
type CatalogChannel struct {
	Name     string            `json:"name"`
	Type     int               `json:"type"`
	BaseURL  string            `json:"baseUrl"`
	Key      string            `json:"key"`
	Models   []string          `json:"models"`
	Group    string            `json:"group"`
	Priority int               `json:"priority"`
	Settings map[string]string `json:"settings,omitempty"`
}

// CatalogModel represents a model entry from the SMS catalog.
type CatalogModel struct {
	ModelID     string  `json:"modelId"`
	DisplayName string  `json:"displayName"`
	Provider    string  `json:"provider"`
	CallType    string  `json:"callType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

// PartitionResponse is the generic wrapper for a partition catalog response.
type PartitionResponse[T any] struct {
	Version int `json:"version"`
	Data    []T `json:"data"`
}
