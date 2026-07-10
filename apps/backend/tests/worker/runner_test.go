package worker_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"

	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestProcessUnknownNewAPISyncOutboxKindFails(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	if err := st.AsyncJobs().EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID: "outbox-unknown", Kind: "unknown_kind", Payload: []byte(`{}`), Status: store.JobStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	entry, found := testutil.NewAPISyncOutboxEntry(st, "outbox-unknown")
	if !found {
		t.Fatal("expected unknown outbox entry to remain in store")
	}
	if entry.Status != store.JobStatusFailed {
		t.Fatalf("expected failed status, got %q", entry.Status)
	}
	if entry.LastError == nil || !strings.Contains(*entry.LastError, "unknown newapi sync outbox kind") {
		t.Fatalf("expected unknown kind error recorded, got %v", entry.LastError)
	}
}

func TestProcessNewAPISyncOutboxDisabledMarksFailed(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, _ := workerfix.NewDisabledNewAPIRunner(t, stub)
	ctx := testutil.Ctx()

	payload, err := json.Marshal(newapisync.UpdateKeyOutboxPayload{
		CompanyID:     contract.DefaultCompanyID,
		PlatformKeyID: contract.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AsyncJobs().EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID: "outbox-newapi-off", Kind: store.OutboxKindUpdateKey, Payload: payload, Status: store.JobStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	entry, found := testutil.NewAPISyncOutboxEntry(st, "outbox-newapi-off")
	if !found {
		t.Fatal("expected outbox entry to remain in store")
	}
	if entry.Status != store.JobStatusFailed {
		t.Fatalf("expected failed status, got %q", entry.Status)
	}
	if pendingNewAPISyncOutbox(st, store.OutboxKindUpdateKey) != 0 {
		t.Fatal("expected no pending update_key outbox after permanent failure")
	}
}

func TestProcessNewAPISyncOutbox(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	runner, st, newAPISync := newWorkerRunner(t, stub)
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

	if err := newAPISync.SyncCreatePlatformKey(ctx, key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}
	if pendingNewAPISyncOutbox(st, store.OutboxKindCreateKey) == 0 {
		t.Fatal("expected pending create_key outbox before RunOnce")
	}

	runner.RunOnce(ctx)

	if stub.CreateTokenCalls < 1 {
		t.Fatalf("expected CreateToken to be called, got %d", stub.CreateTokenCalls)
	}
	if pendingNewAPISyncOutbox(st, store.OutboxKindCreateKey) != 0 {
		t.Fatal("expected newapi sync outbox done after RunOnce")
	}
}

func TestReconcileLogs(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 88, RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	tokenID := int64(88)
	newapisynctf.UpsertMapping(t, st, newapisynctf.MappingOpts{
		PlatformKeyID: contract.IDPlatformKey1, NewAPIKeyID: tokenID,
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

	opts := newapisynctf.DefaultMappingOpts()
	opts.NewAPIKeyID = tokenID
	newapisynctf.UpsertMapping(t, st, opts)

	runner.RunOnce(ctx)

	ingested, err := testutil.HasLedgerLogID(st, logID)
	if err != nil || !ingested {
		t.Fatalf("expected ledger entry after mapping recovery, err=%v ingested=%v", err, ingested)
	}
}
