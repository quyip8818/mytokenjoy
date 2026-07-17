//go:build testhook && integration

package stress_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

// ---------------------------------------------------------------------------
// Stress Environment
// ---------------------------------------------------------------------------

type stressEnvOpts struct {
	KeyBudget       float64
	MemberBudget    float64
	DeptBudget      float64
	AlertThresholds []int
}

type stressEnv struct {
	Cfg      config.Config
	Store    store.Store
	Runtime  *riverfix.TestRuntime
	Ingest   *domainusage.IngestService
	Precheck *gateway.PrecheckService
	Stub     *mock.StubAdminClient
}

func buildStressEnv(t *testing.T, opts stressEnvOpts) *stressEnv {
	t.Helper()

	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 99, RemainQuota: 1_000_000, Key: "sk-stress-99"},
	}
	stub.CreateTokenFn = func(_ context.Context, req newapi.CreateTokenRequest) (newapi.Token, error) {
		return newapi.Token{ID: 99, Key: "sk-stress-99", RemainQuota: 1_000_000}, nil
	}
	stub.UpdateTokenFn = func(_ context.Context, req newapi.UpdateTokenRequest) (newapi.Token, error) {
		return newapi.Token{ID: req.ID, RemainQuota: 0}, nil
	}

	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)

	cfg := runner.Cfg
	precheck := gateway.NewPrecheckServiceLegacy(st.GatewayPrecheck(), cfg.Clock(), nil)

	// Prepare budget fixtures
	ctx := testutil.Ctx()
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	// Set key budget
	if opts.KeyBudget > 0 {
		keys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for i := range keys {
			if keys[i].ID == contract.IDPlatformKey1 {
				keys[i].Budget = opts.KeyBudget
			}
		}
		if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
			t.Fatal(err)
		}
	}

	// Set member budget
	if opts.MemberBudget > 0 {
		if err := st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, opts.MemberBudget); err != nil {
			t.Fatal(err)
		}
	}

	// Set department budget
	if opts.DeptBudget > 0 {
		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !orgfix.SetNodeBudget(nodes, contract.IDDept3, opts.DeptBudget) {
			t.Fatal("dept-3 not found in tree")
		}
		if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
			t.Fatal(err)
		}
	}

	// Set alert rules
	if len(opts.AlertThresholds) > 0 {
		rule := types.AlertRule{
			ID:            "stress-alert-1",
			NodeID:        contract.IDDept3,
			NodeName:      "Dept 3",
			Thresholds:    opts.AlertThresholds,
			NotifyRoleIDs: []string{},
			Enabled:       true,
		}
		if err := st.Budget().SetAlertRules(ctx, []types.AlertRule{rule}); err != nil {
			t.Fatal(err)
		}
	}

	// Reset consumed to 0 so tests start fresh
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1, 0)
	budgetfix.SetMemberSnapshotConsumed(t, st, contract.IDMember1, 0)

	return &stressEnv{
		Cfg:      cfg,
		Store:    st,
		Runtime:  runner,
		Ingest:   ingest,
		Precheck: precheck,
		Stub:     stub,
	}
}

// ---------------------------------------------------------------------------
// Ingest Helpers
// ---------------------------------------------------------------------------

var logIDCounter atomic.Int64

func init() {
	logIDCounter.Store(10000)
}

func nextLogID() int64 {
	return logIDCounter.Add(1)
}

// seedAndIngest creates a consume log and ingests it. Returns the log ID.
func seedAndIngest(t *testing.T, env *stressEnv, quota int64) int64 {
	t.Helper()
	logID := nextLogID()
	raw := store.RawConsumeLog{
		ID:        logID,
		TokenID:   99, // matches default mapping
		Quota:     quota,
		ModelName: "local-test-model",
		CreatedAt: 1781866800, // within clock anchor window
	}
	testutil.SeedConsumeLog(t, env.Store, raw)
	if err := env.Ingest.IngestByLogID(testutil.Ctx(), logID, types.SourceWebhook); err != nil {
		t.Fatalf("ingest logID=%d failed: %v", logID, err)
	}
	return logID
}

// concurrentIngest runs N goroutines each ingesting one log entry.
// Returns all errors (nil entries for success).
func concurrentIngest(t *testing.T, env *stressEnv, n int, quotaPerReq int64) []error {
	t.Helper()
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			logID := nextLogID()
			raw := store.RawConsumeLog{
				ID:        logID,
				TokenID:   99,
				Quota:     quotaPerReq,
				ModelName: "local-test-model",
				CreatedAt: 1781866800,
			}
			testutil.SeedConsumeLog(t, env.Store, raw)
			errs[idx] = env.Ingest.IngestByLogID(testutil.Ctx(), logID, types.SourceWebhook)
		}(i)
	}
	wg.Wait()
	return errs
}

// drainOverrunJobs executes all pending river jobs (including overrun).
func drainOverrunJobs(t *testing.T, env *stressEnv) {
	t.Helper()
	ctx := testutil.Ctx()
	env.Runtime.RunOnce(t, ctx)
}

// ---------------------------------------------------------------------------
// Assertion Helpers
// ---------------------------------------------------------------------------

func assertKeyStatus(t *testing.T, st store.Store, keyID, expectedStatus string) {
	t.Helper()
	ctx := testutil.Ctx()
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatalf("assertKeyStatus: %v", err)
	}
	for _, key := range keys {
		if key.ID == keyID {
			if key.Status != expectedStatus {
				t.Errorf("key %s: expected status=%q, got %q", keyID, expectedStatus, key.Status)
			}
			return
		}
	}
	t.Errorf("key %s not found", keyID)
}

func assertKeyActive(t *testing.T, st store.Store, keyID string) {
	t.Helper()
	assertKeyStatus(t, st, keyID, "active")
}

func assertKeyDisabled(t *testing.T, st store.Store, keyID string) {
	t.Helper()
	ctx := testutil.Ctx()
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatalf("assertKeyDisabled: %v", err)
	}
	for _, key := range keys {
		if key.ID == keyID {
			if key.Status == "active" {
				t.Errorf("key %s: expected disabled/revoked but got active", keyID)
			}
			return
		}
	}
	t.Errorf("key %s not found", keyID)
}

func assertGatewayBlocked(t *testing.T, env *stressEnv, keyHash string) {
	t.Helper()
	// Use a fresh precheck (no cache) to check current PG state
	fresh := gateway.NewPrecheckServiceLegacy(env.Store.GatewayPrecheck(), env.Cfg.Clock(), nil)
	_, err := fresh.Run(testutil.Ctx(), keyHash, "local-test-model", gateway.PrecheckOpts{})
	if err == nil {
		t.Error("expected gateway precheck to block, but it passed")
	}
}

func assertGatewayAllowed(t *testing.T, env *stressEnv, keyHash string) {
	t.Helper()
	// Use a fresh precheck (no cache) to check current PG state
	fresh := gateway.NewPrecheckServiceLegacy(env.Store.GatewayPrecheck(), env.Cfg.Clock(), nil)
	_, err := fresh.Run(testutil.Ctx(), keyHash, "local-test-model", gateway.PrecheckOpts{})
	if err != nil {
		t.Errorf("expected gateway precheck to pass, got: %v", err)
	}
}

// notificationLogsFromPG queries the notification_log table for assertions.
// This captures notifications written by the real notification service via River jobs.
func notificationLogsFromPG(t *testing.T, st store.Store) []types.NotificationLogEntry {
	t.Helper()
	return testutil.NotificationLogs(st)
}

func assertNotificationInPG(t *testing.T, st store.Store, eventType string, minCount int) {
	t.Helper()
	logs := notificationLogsFromPG(t, st)
	var count int
	for _, log := range logs {
		if log.EventType == eventType {
			count++
		}
	}
	if count < minCount {
		t.Errorf("expected at least %d notification(s) of type %q in PG, got %d (total logs: %d)",
			minCount, eventType, count, len(logs))
	}
}

func assertNoNotificationInPG(t *testing.T, st store.Store, eventType string) {
	t.Helper()
	logs := notificationLogsFromPG(t, st)
	for _, log := range logs {
		if log.EventType == eventType {
			t.Errorf("expected no notifications of type %q in PG, but found one", eventType)
			return
		}
	}
}

// keyHashForPlatformKey returns the hash for the default platform key.
func keyHashForPlatformKey(t *testing.T, st store.Store) string {
	t.Helper()
	ctx := testutil.Ctx()
	pool := postgres.MainPool(st)
	var keyHash string
	err := pool.QueryRow(ctx, `SELECT key_hash FROM platform_keys WHERE id = $1`, contract.IDPlatformKey1).Scan(&keyHash)
	if err != nil {
		t.Fatalf("could not get key_hash: %v", err)
	}
	return keyHash
}

// consumedForKey returns the current budget_consumed value for a platform key.
func consumedForKey(t *testing.T, st store.Store, keyID string) float64 {
	t.Helper()
	return budgetfix.PlatformKeySnapshotConsumed(t, st, keyID)
}

// formatPoints formats a float64 as a concise string for logging.
func formatPoints(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
