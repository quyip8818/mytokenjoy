package smssync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/worker/smssync"
)

// --- in-memory model store for testing ---

type storedModel struct {
	ModelID     string
	DisplayName string
	Provider    string
	CallType    string
	InputPrice  float64
	OutputPrice float64
	Source      string
}

type memModelStore struct {
	models map[string]storedModel // keyed by modelID
}

func newMemModelStore() *memModelStore {
	return &memModelStore{models: make(map[string]storedModel)}
}

func (m *memModelStore) DisableStaleFromSMS(_ context.Context, activeIDs []string) (int, error) {
	active := make(map[string]bool, len(activeIDs))
	for _, id := range activeIDs { active[id] = true }
	var count int
	for k := range m.models { if !active[k] { delete(m.models, k); count++ } }
	return count, nil
}

func (m *memModelStore) UpsertFromSMS(_ context.Context, model sms.CatalogModel) error {
	m.models[model.ModelID] = storedModel{
		ModelID:     model.ModelID,
		DisplayName: model.DisplayName,
		Provider:    model.Provider,
		CallType:    model.CallType,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		Source:      "sms",
	}
	return nil
}

func TestAdminPortTarget_UpsertModel_WritesToStore(t *testing.T) {
	store := newMemModelStore()
	// Use nil for adminport.Port since we're only testing UpsertModel path
	target := smssync.NewAdminPortTarget(nil, store)

	model := sms.CatalogModel{
		ModelID:     "deepseek-chat",
		DisplayName: "DeepSeek Chat V3",
		Provider:    "deepseek",
		CallType:    "chat",
		InputPrice:  1.0,
		OutputPrice: 2.0,
	}

	err := target.UpsertModel(context.Background(), model)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stored, ok := store.models["deepseek-chat"]
	if !ok {
		t.Fatal("expected model to be stored")
	}
	if stored.Source != "sms" {
		t.Fatalf("expected source=sms, got %s", stored.Source)
	}
	if stored.DisplayName != "DeepSeek Chat V3" {
		t.Fatalf("expected display name DeepSeek Chat V3, got %s", stored.DisplayName)
	}
	if stored.InputPrice != 1.0 {
		t.Fatalf("expected input price 1.0, got %f", stored.InputPrice)
	}
}
