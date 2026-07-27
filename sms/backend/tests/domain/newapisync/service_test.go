package newapisync_test

import (
	"context"
	"testing"

	"log/slog"

	"sms/backend/internal/domain/newapisync"
	"sms/backend/internal/domain/types"
)

// --- mock admin port ---

type mockAdmin struct {
	ratios  map[string]newapisync.ModelPricing
	synced  []newapisync.PricingEntry
	upserts []struct {
		ModelID     string
		InputPrice  float64
		OutputPrice float64
	}
}

func newMockAdmin() *mockAdmin {
	return &mockAdmin{ratios: make(map[string]newapisync.ModelPricing)}
}

func (m *mockAdmin) ListCurrentRatios(_ context.Context) (map[string]newapisync.ModelPricing, error) {
	return m.ratios, nil
}

func (m *mockAdmin) SyncPricing(_ context.Context, entries []newapisync.PricingEntry) error {
	m.synced = entries
	return nil
}

func (m *mockAdmin) UpsertModelRatio(_ context.Context, modelID string, inputPrice, outputPrice float64) error {
	m.upserts = append(m.upserts, struct {
		ModelID     string
		InputPrice  float64
		OutputPrice float64
	}{modelID, inputPrice, outputPrice})
	return nil
}

func (m *mockAdmin) ListModels(_ context.Context) ([]newapisync.NewAPIModel, error) {

func (m *mockAdmin) ListChannels(_ context.Context) ([]newapisync.Channel, error) {
	return nil, nil
}
	var models []newapisync.NewAPIModel
	for id, p := range m.ratios {
		models = append(models, newapisync.NewAPIModel{
			ModelID:         id,
			InputPrice:      p.InputPrice,
			OutputPrice:     p.OutputPrice,
			ModelRatio:      p.ModelRatio,
			CompletionRatio: p.CompletionRatio,
		})
	}
	return models, nil
}

// --- mock model lister ---

type mockModelLister struct {
	models []types.AiModel
}

func (m *mockModelLister) ListWithModelID(_ context.Context) ([]types.AiModel, error) {
	return m.models, nil
}

// --- helpers ---

func ptr[T any](v T) *T { return &v }

func newTestService(admin *mockAdmin, models []types.AiModel) *newapisync.Service {
	lister := &mockModelLister{models: models}
	return newapisync.NewService(admin, lister, slog.Default())
}

// --- tests ---

func TestGetStatus_Synced(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	admin.ratios["gpt-4o"] = newapisync.ModelPricing{
		ModelID: "gpt-4o", ModelRatio: 30, CompletionRatio: 2,
		InputPrice: 60, OutputPrice: 120,
	}
	models := []types.AiModel{
		{ModelName: "GPT-4o", ModelID: ptr("gpt-4o"), InputPrice: ptr(60.0), OutputPrice: ptr(120.0)},
	}
	svc := newTestService(admin, models)

	statuses, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Status != "synced" {
		t.Fatalf("expected synced, got %s", statuses[0].Status)
	}
}

func TestGetStatus_Diverged(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	admin.ratios["gpt-4o"] = newapisync.ModelPricing{
		ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 100, // different output
	}
	models := []types.AiModel{
		{ModelName: "GPT-4o", ModelID: ptr("gpt-4o"), InputPrice: ptr(60.0), OutputPrice: ptr(120.0)},
	}
	svc := newTestService(admin, models)

	statuses, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if statuses[0].Status != "diverged" {
		t.Fatalf("expected diverged, got %s", statuses[0].Status)
	}
}

func TestGetStatus_Missing(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin() // empty ratios
	models := []types.AiModel{
		{ModelName: "GPT-4o", ModelID: ptr("gpt-4o"), InputPrice: ptr(60.0), OutputPrice: ptr(120.0)},
	}
	svc := newTestService(admin, models)

	statuses, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if statuses[0].Status != "missing" {
		t.Fatalf("expected missing, got %s", statuses[0].Status)
	}
}

func TestSyncAll_FiltersCorrectly(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	models := []types.AiModel{
		{ModelName: "A", ModelID: ptr("model-a"), InputPrice: ptr(60.0), OutputPrice: ptr(120.0)},
		{ModelName: "B", ModelID: ptr("model-b"), InputPrice: ptr(0.0)}, // no price, skipped
		{ModelName: "C", ModelID: nil, InputPrice: ptr(10.0)},           // no model_id, skipped
		{ModelName: "D", ModelID: ptr("model-d"), InputPrice: ptr(20.0), OutputPrice: ptr(40.0)},
	}
	svc := newTestService(admin, models)

	count, err := svc.SyncAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 synced, got %d", count)
	}
	if len(admin.synced) != 2 {
		t.Fatalf("expected 2 entries passed to SyncPricing, got %d", len(admin.synced))
	}
	// Verify entries
	if admin.synced[0].ModelID != "model-a" {
		t.Fatalf("expected model-a, got %s", admin.synced[0].ModelID)
	}
	if admin.synced[1].ModelID != "model-d" {
		t.Fatalf("expected model-d, got %s", admin.synced[1].ModelID)
	}
}

func TestSyncAll_EmptyList(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	svc := newTestService(admin, nil)

	count, err := svc.SyncAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
	if admin.synced != nil {
		t.Fatal("expected no SyncPricing call")
	}
}

func TestUpsertOne_SkipsEmptyModelID(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	svc := newTestService(admin, nil)

	svc.UpsertOne(context.Background(), "", 60, 120)
	if len(admin.upserts) != 0 {
		t.Fatal("expected no upsert for empty modelID")
	}
}

func TestUpsertOne_SkipsZeroPrice(t *testing.T) {
	t.Parallel()
	admin := newMockAdmin()
	svc := newTestService(admin, nil)

	svc.UpsertOne(context.Background(), "gpt-4o", 0, 120)
	if len(admin.upserts) != 0 {
		t.Fatal("expected no upsert for zero input price")
	}
}
