package smssync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/worker/smssync"
)

// --- mock SMS client ---

type mockSMSClient struct {
	catalog *sms.Catalog
	err     error
}

func (m *mockSMSClient) FetchCatalog(_ context.Context) (*sms.Catalog, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.catalog, nil
}

// --- mock sync target ---

type syncCall struct {
	method string
	args   interface{}
}

type mockSyncTarget struct {
	calls []syncCall
}

func (m *mockSyncTarget) UpsertChannel(_ context.Context, ch sms.CatalogChannel) error {
	m.calls = append(m.calls, syncCall{"UpsertChannel", ch})
	return nil
}

func (m *mockSyncTarget) UpsertModelRatio(_ context.Context, modelID string, inputPrice, outputPrice float64) error {
	m.calls = append(m.calls, syncCall{"UpsertModelRatio", []interface{}{modelID, inputPrice, outputPrice}})
	return nil
}

func (m *mockSyncTarget) UpsertModel(_ context.Context, model sms.CatalogModel) error {
	m.calls = append(m.calls, syncCall{"UpsertModel", model})
	return nil
}

func (m *mockSyncTarget) RebuildAbilities(_ context.Context) error {
	m.calls = append(m.calls, syncCall{"RebuildAbilities", nil})
	return nil
}

// --- tests ---

func TestSyncWorker_Execute_FullSync(t *testing.T) {
	client := &mockSMSClient{
		catalog: &sms.Catalog{
			Channels: []sms.CatalogChannel{
				{Name: "deepseek", Type: 43, BaseURL: "https://api.deepseek.com", Key: "sk-x", Models: []string{"deepseek-chat"}, Group: "default"},
			},
			Models: []sms.CatalogModel{
				{ModelID: "deepseek-chat", DisplayName: "DeepSeek Chat", Provider: "deepseek", CallType: "chat", InputPrice: 1.0, OutputPrice: 2.0},
			},
			SyncedAt: time.Now(),
		},
	}
	target := &mockSyncTarget{}
	worker := smssync.New(client, target)

	err := worker.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have called: UpsertChannel, RebuildAbilities, UpsertModelRatio, UpsertModel
	if len(target.calls) != 4 {
		t.Fatalf("expected 4 calls, got %d: %+v", len(target.calls), target.calls)
	}
	if target.calls[0].method != "UpsertChannel" {
		t.Fatalf("expected first call UpsertChannel, got %s", target.calls[0].method)
	}
	if target.calls[1].method != "RebuildAbilities" {
		t.Fatalf("expected second call RebuildAbilities, got %s", target.calls[1].method)
	}
	if target.calls[2].method != "UpsertModelRatio" {
		t.Fatalf("expected third call UpsertModelRatio, got %s", target.calls[2].method)
	}
	if target.calls[3].method != "UpsertModel" {
		t.Fatalf("expected fourth call UpsertModel, got %s", target.calls[3].method)
	}
}

func TestSyncWorker_Execute_SMSUnreachable(t *testing.T) {
	client := &mockSMSClient{err: errors.New("connection refused")}
	target := &mockSyncTarget{}
	worker := smssync.New(client, target)

	err := worker.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error when SMS unreachable")
	}
	// Should not have called any target methods
	if len(target.calls) != 0 {
		t.Fatalf("expected 0 target calls, got %d", len(target.calls))
	}
}
