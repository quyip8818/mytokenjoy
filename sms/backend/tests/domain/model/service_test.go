package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"sms/backend/internal/domain/model"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	models []types.AiModel
}

func newMockStore() *mockStore {
	return &mockStore{}
}

func (m *mockStore) ListModels(_ context.Context, f model.ListFilter) (*types.PagedResult[types.AiModel], error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 10
	}
	return &types.PagedResult[types.AiModel]{
		Items: m.models, Total: len(m.models), Page: page, PageSize: size,
	}, nil
}

func (m *mockStore) GetModel(_ context.Context, id uuid.UUID) (*types.AiModel, error) {
	for i := range m.models {
		if m.models[i].ID == id {
			return &m.models[i], nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateModel(_ context.Context, md *types.AiModel) error {
	md.ID = uuid.Must(uuid.NewV7())
	m.models = append(m.models, *md)
	return nil
}

func (m *mockStore) UpdateModel(_ context.Context, id uuid.UUID, md *types.AiModel) error {
	for i := range m.models {
		if m.models[i].ID == id {
			md.ID = id
			m.models[i] = *md
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) DeleteModel(_ context.Context, id uuid.UUID) error {
	for i := range m.models {
		if m.models[i].ID == id {
			m.models = append(m.models[:i], m.models[i+1:]...)
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStore) ListModelsWithModelID(_ context.Context) ([]types.AiModel, error) {
	var result []types.AiModel
	for _, md := range m.models {
		if md.ModelID != nil && *md.ModelID != "" {
			result = append(result, md)
		}
	}
	return result, nil
}

// --- mock syncer ---

type mockSyncer struct {
	calls []syncerCall
}

type syncerCall struct {
	ModelID     string
	InputPrice  float64
	OutputPrice float64
}

func (ms *mockSyncer) UpsertOne(_ context.Context, modelID string, inputPrice, outputPrice float64) {
	ms.calls = append(ms.calls, syncerCall{modelID, inputPrice, outputPrice})
}

// --- tests ---

var testSupplierID = uuid.Must(uuid.NewV7())

func newService() *model.Service {
	return model.NewService(newMockStore())
}

func newServiceWithSyncer() (*model.Service, *mockSyncer) {
	svc := model.NewService(newMockStore())
	ms := &mockSyncer{}
	svc.SetSyncer(ms)
	return svc, ms
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	m, err := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID, ModelName: "GPT-4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.ID == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if m.Status != "available" {
		t.Fatalf("expected default status available, got %s", m.Status)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := newService()
	cases := []struct {
		name  string
		input model.CreateInput
	}{
		{"empty name", model.CreateInput{SupplierID: &testSupplierID}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID, ModelName: "X", Status: "bad",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID, ModelName: "Claude",
	})
	updated, err := svc.Update(context.Background(), created.ID, model.UpdateInput{
		SupplierID: &testSupplierID, ModelName: "Claude-3", Status: "deprecated",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ModelName != "Claude-3" {
		t.Fatalf("expected Claude-3, got %s", updated.ModelName)
	}
	if updated.Status != "deprecated" {
		t.Fatalf("expected deprecated, got %s", updated.Status)
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID, ModelName: "X",
	})
	_, err := svc.Update(context.Background(), created.ID, model.UpdateInput{
		SupplierID: &testSupplierID, ModelName: "X", Status: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID, ModelName: "ToDelete",
	})
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Get(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestList_CapsPageSize(t *testing.T) {
	t.Parallel()
	svc := newService()
	result, err := svc.List(context.Background(), model.ListFilter{Page: 1, PageSize: 200})
	if err != nil {
		t.Fatal(err)
	}
	if result.PageSize > 100 {
		t.Fatalf("expected pageSize capped to 100, got %d", result.PageSize)
	}
}

func ptr[T any](v T) *T { return &v }

func TestCreate_TriggersSyncWhenModelIDAndPrice(t *testing.T) {
	t.Parallel()
	svc, ms := newServiceWithSyncer()
	_, err := svc.Create(context.Background(), model.CreateInput{
		SupplierID:  &testSupplierID,
		ModelName:   "GPT-4o",
		ModelID:     ptr("gpt-4o"),
		InputPrice:  ptr(60.0),
		OutputPrice: ptr(120.0),
	})
	if err != nil {
		t.Fatal(err)
	}
	// triggerSync fires a goroutine; give it a moment
	waitForSyncer(ms, 1)
	if len(ms.calls) != 1 {
		t.Fatalf("expected 1 sync call, got %d", len(ms.calls))
	}
	if ms.calls[0].ModelID != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", ms.calls[0].ModelID)
	}
	if ms.calls[0].InputPrice != 60 || ms.calls[0].OutputPrice != 120 {
		t.Fatalf("unexpected prices: %+v", ms.calls[0])
	}
}

func TestCreate_NoSyncWithoutModelID(t *testing.T) {
	t.Parallel()
	svc, ms := newServiceWithSyncer()
	_, err := svc.Create(context.Background(), model.CreateInput{
		SupplierID:  &testSupplierID,
		ModelName:   "NoID",
		InputPrice:  ptr(60.0),
		OutputPrice: ptr(120.0),
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForSyncer(ms, 0)
	if len(ms.calls) != 0 {
		t.Fatalf("expected no sync call, got %d", len(ms.calls))
	}
}

func TestCreate_NoSyncWithZeroPrice(t *testing.T) {
	t.Parallel()
	svc, ms := newServiceWithSyncer()
	_, err := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID,
		ModelName:  "ZeroPrice",
		ModelID:    ptr("zero-model"),
		// no InputPrice set (nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForSyncer(ms, 0)
	if len(ms.calls) != 0 {
		t.Fatalf("expected no sync for nil price, got %d", len(ms.calls))
	}
}

func TestUpdate_TriggersSync(t *testing.T) {
	t.Parallel()
	svc, ms := newServiceWithSyncer()
	created, _ := svc.Create(context.Background(), model.CreateInput{
		SupplierID: &testSupplierID,
		ModelName:  "Claude",
	})
	// clear create calls (no model_id, so should be 0 anyway)
	ms.calls = nil

	_, err := svc.Update(context.Background(), created.ID, model.UpdateInput{
		SupplierID:  &testSupplierID,
		ModelName:   "Claude",
		ModelID:     ptr("claude-3-5"),
		InputPrice:  ptr(45.0),
		OutputPrice: ptr(90.0),
		Status:      "available",
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForSyncer(ms, 1)
	if len(ms.calls) != 1 {
		t.Fatalf("expected 1 sync call after update, got %d", len(ms.calls))
	}
	if ms.calls[0].ModelID != "claude-3-5" {
		t.Fatalf("expected claude-3-5, got %s", ms.calls[0].ModelID)
	}
}

// waitForSyncer polls briefly for expected call count (triggerSync uses goroutine)
func waitForSyncer(ms *mockSyncer, expected int) {
	for i := 0; i < 100; i++ {
		if len(ms.calls) >= expected {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
