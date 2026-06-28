package worker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestProcessRelayOutbox(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, lifecycle := newWorkerRunner(t, stub)
	ctx := context.Background()

	memberID := seed.IDMember1
	key := types.PlatformKey{
		ID: "plk-worker", Name: "worker-key", MemberID: &memberID,
		Status: "active", Quota: 1000, ModelWhitelist: []string{"gpt-4o"},
	}
	keys := st.Keys().PlatformKeys()
	keys = append(keys, key)
	st.Keys().SetPlatformKeys(keys)

	if err := lifecycle.SyncCreatePlatformKey(ctx, key, seed.IDDept3); err != nil {
		t.Fatal(err)
	}
	if pendingRelayOutbox(st, store.OutboxKindCreateToken) == 0 {
		t.Fatal("expected pending create_token outbox before RunOnce")
	}

	runner.RunOnce(ctx)

	if stub.CreateTokenCalls < 1 {
		t.Fatalf("expected CreateToken to be called, got %d", stub.CreateTokenCalls)
	}
	if pendingRelayOutbox(st, store.OutboxKindCreateToken) != 0 {
		t.Fatal("expected relay outbox done after RunOnce")
	}
}

func TestProcessWebhookOutbox(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 77, RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := context.Background()

	tokenID := int64(77)
	memberID := seed.IDMember1
	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      &memberID,
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(newapi.WebhookLogPayload{
		ID: 1001, TokenID: 77, Quota: 500, Model: "gpt-4o", CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Relay().EnqueueWebhookOutbox(store.WebhookOutboxEntry{
		ID: "wh-1", Payload: payload, Status: store.OutboxStatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	if pendingWebhookOutbox(st) == 0 {
		t.Fatal("expected pending webhook outbox before RunOnce")
	}

	runner.RunOnce(ctx)

	if pendingWebhookOutbox(st) != 0 {
		t.Fatal("expected webhook outbox done after RunOnce")
	}
	ingested, err := st.Relay().HasIngestedLogID(1001)
	if err != nil || !ingested {
		t.Fatalf("expected log 1001 ingested, err=%v ingested=%v", err, ingested)
	}
}

func TestCompensateLogs(t *testing.T) {
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 88, RemainQuota: 1000},
		ListLogsFn: func(_ context.Context, params newapi.ListLogsParams) ([]newapi.LogEntry, error) {
			logs := []newapi.LogEntry{
				{ID: 500, TokenID: 88, Quota: 300, ModelName: "gpt-4o", CreatedAt: time.Now().Unix()},
			}
			out := make([]newapi.LogEntry, 0)
			for _, entry := range logs {
				if entry.ID > params.StartID {
					out = append(out, entry)
				}
			}
			return out, nil
		},
	}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := context.Background()

	tokenID := int64(88)
	memberID := seed.IDMember1
	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      &memberID,
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Relay().SetLastLogID(0); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	ingested, err := st.Relay().HasIngestedLogID(500)
	if err != nil || !ingested {
		t.Fatalf("expected log 500 ingested via compensation, err=%v ingested=%v", err, ingested)
	}
}
