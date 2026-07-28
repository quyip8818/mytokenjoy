package sync

import (
	"context"
	"time"
)

type CatalogModel struct {
	ModelID     string  `json:"modelId"`
	DisplayName string  `json:"displayName"`
	Provider    string  `json:"provider"`
	CallType    string  `json:"callType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

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

// PartitionVersions holds the current version of each sync partition.
type PartitionVersions struct {
	Channels   int `json:"channels"`
	Models     int `json:"models"`
	Currencies int `json:"currencies"`
}

// PartitionResponse wraps a partition's version + data array.
type PartitionResponse[T any] struct {
	Version int `json:"version"`
	Data    []T `json:"data"`
}

// Catalog is the legacy full-catalog response (kept for backward compat).
type Catalog struct {
	Channels []CatalogChannel `json:"channels"`
	Models   []CatalogModel   `json:"models"`
	SyncedAt time.Time        `json:"syncedAt"`
}

type Store interface {
	ListModelsForSync(ctx context.Context) ([]CatalogModel, error)
	ListChannelsForSync(ctx context.Context) ([]CatalogChannel, error)
	GetPartitionVersions(ctx context.Context) (PartitionVersions, error)
	GetPartitionVersion(ctx context.Context, partition string) (int, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// GetVersions returns the current version of each partition.
func (s *Service) GetVersions(ctx context.Context) (PartitionVersions, error) {
	return s.store.GetPartitionVersions(ctx)
}

// GetModels returns the models partition catalog with version.
func (s *Service) GetModels(ctx context.Context) (*PartitionResponse[CatalogModel], error) {
	version, err := s.store.GetPartitionVersion(ctx, "models")
	if err != nil {
		return nil, err
	}
	models, err := s.store.ListModelsForSync(ctx)
	if err != nil {
		return nil, err
	}
	if models == nil {
		models = []CatalogModel{}
	}
	return &PartitionResponse[CatalogModel]{Version: version, Data: models}, nil
}

// GetChannels returns the channels partition catalog with version.
func (s *Service) GetChannels(ctx context.Context) (*PartitionResponse[CatalogChannel], error) {
	version, err := s.store.GetPartitionVersion(ctx, "channels")
	if err != nil {
		return nil, err
	}
	channels, err := s.store.ListChannelsForSync(ctx)
	if err != nil {
		return nil, err
	}
	if channels == nil {
		channels = []CatalogChannel{}
	}
	return &PartitionResponse[CatalogChannel]{Version: version, Data: channels}, nil
}

// GetCatalog returns the legacy full catalog (backward compat with old endpoint).
func (s *Service) GetCatalog(ctx context.Context) (Catalog, error) {
	models, err := s.store.ListModelsForSync(ctx)
	if err != nil {
		return Catalog{}, err
	}
	channels, err := s.store.ListChannelsForSync(ctx)
	if err != nil {
		return Catalog{}, err
	}
	if models == nil {
		models = []CatalogModel{}
	}
	if channels == nil {
		channels = []CatalogChannel{}
	}
	return Catalog{
		Models:   models,
		Channels: channels,
		SyncedAt: time.Now(),
	}, nil
}
