package sync_test

import (
	"context"
	"testing"

	"sms/backend/internal/domain/sync"
)

// --- mock store ---

type mockModelStore struct{}

func (m *mockModelStore) ListModelsForSync(_ context.Context) ([]sync.CatalogModel, error) {
	return []sync.CatalogModel{
		{
			ModelID:     "deepseek-chat",
			DisplayName: "DeepSeek Chat V3",
			Provider:    "deepseek",
			CallType:    "chat",
			InputPrice:  1.0,
			OutputPrice: 2.0,
		},
		{
			ModelID:     "gpt-4o",
			DisplayName: "GPT-4o",
			Provider:    "openai",
			CallType:    "chat",
			InputPrice:  5.0,
			OutputPrice: 15.0,
		},
	}, nil
}

func (m *mockModelStore) ListChannelsForSync(_ context.Context) ([]sync.CatalogChannel, error) {
	return []sync.CatalogChannel{
		{
			Name:     "deepseek-official",
			Type:     43,
			BaseURL:  "https://api.deepseek.com",
			Key:      "sk-xxx",
			Models:   []string{"deepseek-chat"},
			Group:    "default",
			Priority: 0,
		},
	}, nil
}

func (m *mockModelStore) GetPartitionVersions(_ context.Context) (sync.PartitionVersions, error) {
	return sync.PartitionVersions{Channels: 1, Models: 2, Currencies: 0}, nil
}

func (m *mockModelStore) GetPartitionVersion(_ context.Context, _ string) (int, error) {
	return 1, nil
}

// --- tests ---

func TestGetCatalog_ReturnsModelsAndChannels(t *testing.T) {
	store := &mockModelStore{}
	svc := sync.NewService(store)

	catalog, err := svc.GetCatalog(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(catalog.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(catalog.Models))
	}
	if catalog.Models[0].ModelID != "deepseek-chat" {
		t.Fatalf("expected first model deepseek-chat, got %s", catalog.Models[0].ModelID)
	}
	if catalog.Models[0].InputPrice != 1.0 {
		t.Fatalf("expected input price 1.0, got %f", catalog.Models[0].InputPrice)
	}

	if len(catalog.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(catalog.Channels))
	}
	if catalog.Channels[0].Name != "deepseek-official" {
		t.Fatalf("expected channel deepseek-official, got %s", catalog.Channels[0].Name)
	}

	if catalog.SyncedAt.IsZero() {
		t.Fatal("expected non-zero syncedAt")
	}
}
