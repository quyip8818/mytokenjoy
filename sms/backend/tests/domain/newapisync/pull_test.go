package newapisync_test

import (
	"context"
	"testing"

	"log/slog"

	"sms/backend/internal/domain/newapisync"
)

// --- mock pull store ---

type mockPullStore struct {
	channels        []newapisync.Channel
	models          []newapisync.SyncedModelInput
	deprecatedIDs   []string
	deprecatedCount int
}

func (m *mockPullStore) UpsertChannel(_ context.Context, ch *newapisync.Channel) error {
	m.channels = append(m.channels, *ch)
	return nil
}

func (m *mockPullStore) UpsertSyncedModel(_ context.Context, model *newapisync.SyncedModelInput) error {
	m.models = append(m.models, *model)
	return nil
}

func (m *mockPullStore) DeprecateStaleSyncModels(_ context.Context, activeIDs []string) (int, error) {
	m.deprecatedIDs = activeIDs
	return m.deprecatedCount, nil
}

// --- mock admin with channels ---

type mockAdminWithChannels struct {
	mockAdmin
	channels []newapisync.Channel
}

func (m *mockAdminWithChannels) ListChannels(_ context.Context) ([]newapisync.Channel, error) {
	return m.channels, nil
}

// --- helpers ---

func newPullTestService(admin newapisync.AdminPort, store newapisync.PullStore) *newapisync.Service {
	return newapisync.NewServiceWithPullStore(admin, nil, store, slog.Default())
}

// --- tests ---

func TestPullFromNewAPI_SyncsChannelsAndModels(t *testing.T) {
	t.Parallel()
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{
			"gpt-4o":        {ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 120},
			"claude-sonnet": {ModelID: "claude-sonnet", InputPrice: 30, OutputPrice: 90},
		}},
		channels: []newapisync.Channel{
			{ID: 1, Name: "OpenAI", Type: 1, Status: 1, Models: "gpt-4o"},
			{ID: 2, Name: "Anthropic", Type: 14, Status: 1, Models: "claude-sonnet"},
		},
	}
	store := &mockPullStore{}
	svc := newPullTestService(admin, store)

	result, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Channels synced
	if result.ChannelsSynced != 2 {
		t.Fatalf("expected 2 channels synced, got %d", result.ChannelsSynced)
	}
	if len(store.channels) != 2 {
		t.Fatalf("expected 2 channels in store, got %d", len(store.channels))
	}
	if store.channels[0].Name != "OpenAI" {
		t.Fatalf("expected OpenAI, got %s", store.channels[0].Name)
	}

	// Models synced
	if result.ModelsCreated != 2 {
		t.Fatalf("expected 2 models created, got %d", result.ModelsCreated)
	}
	if len(store.models) != 2 {
		t.Fatalf("expected 2 models in store, got %d", len(store.models))
	}
}

func TestPullFromNewAPI_ModelGetsCostFromRatios(t *testing.T) {
	t.Parallel()
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{
			"gpt-4o": {ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 120},
		}},
		channels: []newapisync.Channel{
			{ID: 1, Name: "OpenAI", Type: 1, Status: 1, Models: "gpt-4o"},
		},
	}
	store := &mockPullStore{}
	svc := newPullTestService(admin, store)

	_, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(store.models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(store.models))
	}
	m := store.models[0]
	if m.ModelID != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", m.ModelID)
	}
	if m.CostInput != 60 {
		t.Fatalf("expected cost input 60, got %f", m.CostInput)
	}
	if m.CostOutput != 120 {
		t.Fatalf("expected cost output 120, got %f", m.CostOutput)
	}
	if m.ChannelID != 1 {
		t.Fatalf("expected channel ID 1, got %d", m.ChannelID)
	}
}

func TestPullFromNewAPI_MultipleModelsPerChannel(t *testing.T) {
	t.Parallel()
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{
			"gpt-4o":      {ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 120},
			"gpt-4o-mini": {ModelID: "gpt-4o-mini", InputPrice: 10, OutputPrice: 20},
		}},
		channels: []newapisync.Channel{
			{ID: 1, Name: "OpenAI", Type: 1, Status: 1, Models: "gpt-4o,gpt-4o-mini"},
		},
	}
	store := &mockPullStore{}
	svc := newPullTestService(admin, store)

	result, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if result.ChannelsSynced != 1 {
		t.Fatalf("expected 1 channel, got %d", result.ChannelsSynced)
	}
	if result.ModelsCreated != 2 {
		t.Fatalf("expected 2 models, got %d", result.ModelsCreated)
	}
}

func TestPullFromNewAPI_DeprecatesStaleModels(t *testing.T) {
	t.Parallel()
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{
			"gpt-4o": {ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 120},
		}},
		channels: []newapisync.Channel{
			{ID: 1, Name: "OpenAI", Type: 1, Status: 1, Models: "gpt-4o"},
		},
	}
	store := &mockPullStore{deprecatedCount: 3}
	svc := newPullTestService(admin, store)

	result, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if result.ModelsRemoved != 3 {
		t.Fatalf("expected 3 removed, got %d", result.ModelsRemoved)
	}
	// Verify active IDs passed to deprecate
	if len(store.deprecatedIDs) != 1 || store.deprecatedIDs[0] != "gpt-4o" {
		t.Fatalf("expected active IDs [gpt-4o], got %v", store.deprecatedIDs)
	}
}

func TestPullFromNewAPI_EmptyChannels(t *testing.T) {
	t.Parallel()
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{}},
		channels:  []newapisync.Channel{},
	}
	store := &mockPullStore{}
	svc := newPullTestService(admin, store)

	result, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if result.ChannelsSynced != 0 {
		t.Fatalf("expected 0 channels, got %d", result.ChannelsSynced)
	}
	if result.ModelsCreated != 0 {
		t.Fatalf("expected 0 models, got %d", result.ModelsCreated)
	}
}

func TestPullFromNewAPI_ModelWithoutPricingStillSynced(t *testing.T) {
	t.Parallel()
	// Channel has a model that's not in the ratio map (no pricing configured yet)
	admin := &mockAdminWithChannels{
		mockAdmin: mockAdmin{ratios: map[string]newapisync.ModelPricing{}},
		channels: []newapisync.Channel{
			{ID: 1, Name: "OpenAI", Type: 1, Status: 1, Models: "gpt-4o"},
		},
	}
	store := &mockPullStore{}
	svc := newPullTestService(admin, store)

	result, err := svc.PullFromNewAPI(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Model should still be created (with zero cost)
	if result.ModelsCreated != 1 {
		t.Fatalf("expected 1 model, got %d", result.ModelsCreated)
	}
	if store.models[0].CostInput != 0 {
		t.Fatalf("expected 0 cost, got %f", store.models[0].CostInput)
	}
}
