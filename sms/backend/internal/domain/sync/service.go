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

type Catalog struct {
	Channels []CatalogChannel `json:"channels"`
	Models   []CatalogModel   `json:"models"`
	SyncedAt time.Time        `json:"syncedAt"`
}

type Store interface {
	ListModelsForSync(ctx context.Context) ([]CatalogModel, error)
	ListChannelsForSync(ctx context.Context) ([]CatalogChannel, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) GetCatalog(ctx context.Context) (Catalog, error) {
	models, err := s.store.ListModelsForSync(ctx)
	if err != nil {
		return Catalog{}, err
	}
	channels, err := s.store.ListChannelsForSync(ctx)
	if err != nil {
		return Catalog{}, err
	}
	return Catalog{
		Models:   models,
		Channels: channels,
		SyncedAt: time.Now(),
	}, nil
}
