package worker_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"

	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestProcessUnknownRelayOutboxKindFails(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	if err := st.Relay().EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID: "outbox-unknown", Kind: "unknown_kind", Payload: []byte(`{}`), Status: store.OutboxStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	entry, found := testutil.RelayOutboxEntry(st, "outbox-unknown")
	if !found {
		t.Fatal("expected unknown outbox entry to remain in store")
	}
	if entry.Status != store.OutboxStatusFailed {
		t.Fatalf("expected failed status, got %q", entry.Status)
	}
	if entry.LastError == nil || !strings.Contains(*entry.LastError, "unknown relay outbox kind") {
		t.Fatalf("expected unknown kind error recorded, got %v", entry.LastError)
	}
}

func TestProcessRelayOutboxRelayDisabledMarksFailed(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, _ := workerfix.NewRelayDisabledRunner(t, stub)
	ctx := testutil.Ctx()

	payload, err := json.Marshal(relay.UpdateTokenOutboxPayload{
		CompanyID:     contract.DefaultCompanyID,
		PlatformKeyID: contract.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Relay().EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID: "outbox-relay-off", Kind: store.OutboxKindUpdateToken, Payload: payload, Status: store.OutboxStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	entry, found := testutil.RelayOutboxEntry(st, "outbox-relay-off")
	if !found {
		t.Fatal("expected outbox entry to remain in store")
	}
	if entry.Status != store.OutboxStatusFailed {
		t.Fatalf("expected failed status, got %q", entry.Status)
	}
	if pendingRelayOutbox(st, store.OutboxKindUpdateToken) != 0 {
		t.Fatal("expected no pending update_token outbox after permanent failure")
	}
}

func TestProcessRelayOutbox(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, lifecycle := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-worker", Name: "worker-key", MemberID: &memberID,
		Status: "active", Quota: 1000, ModelWhitelist: []int64{contract.IDModel1},
		CreatedAt: "2026-06-19",
	}
	keys, err := st.Keys().PlatformKeys(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := st.Keys().SetPlatformKeys(testutil.Ctx(), keys); err != nil {
		t.Fatal(err)
	}

	if err := lifecycle.SyncCreatePlatformKey(ctx, key, contract.IDDept3); err != nil {
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

func TestReconcileLogs(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 88, RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	tokenID := int64(88)
	relayfix.UpsertMapping(t, st, relayfix.MappingOpts{
		PlatformKeyID: contract.IDPlatformKey1, NewAPITokenID: tokenID,
	})
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(500, tokenID))

	if err := runner.RunReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}

	ingested, err := testutil.HasLedgerLogID(st, 500)
	if err != nil || !ingested {
		t.Fatalf("expected log 500 in ledger via reconcile, err=%v ingested=%v", err, ingested)
	}
	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil {
		t.Fatal(err)
	}
	if cursor != 500 {
		t.Fatalf("expected cursor 500, got %d", cursor)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestIngestJobMappingLateRecovery(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	const logID = int64(601)
	const tokenID = int64(77)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(logID, tokenID))

	if err := st.Logs().UpsertJob(ctx, store.IngestJobFromError(logID, types.SourceWebhook, errTest("mapping not found"))); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)
	if err := st.Logs().MarkJobRetry(ctx, store.IngestJobID(logID), -time.Second, "mapping not found"); err != nil {
		t.Fatal(err)
	}

	opts := relayfix.DefaultMappingOpts()
	opts.NewAPITokenID = tokenID
	relayfix.UpsertMapping(t, st, opts)

	runner.RunOnce(ctx)

	ingested, err := testutil.HasLedgerLogID(st, logID)
	if err != nil || !ingested {
		t.Fatalf("expected ledger entry after mapping recovery, err=%v ingested=%v", err, ingested)
	}
}
