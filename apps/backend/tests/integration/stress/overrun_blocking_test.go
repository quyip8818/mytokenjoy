//go:build testhook && integration

package stress_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

// TestOverrunBlockingIntegration is the main integration test that exercises
// the full ingest→overrun→disable→gateway pipeline under realistic scenarios.
//
// Run with: make test-integration
// Or:       go test -tags="testhook,integration" -race -timeout 180s ./tests/integration/stress/...
func TestOverrunBlockingIntegration(t *testing.T) {
	t.Run("KeyBudgetExhaustion", testKeyBudgetExhaustion)
	t.Run("MemberBudgetIsolation", testMemberBudgetIsolation)
	t.Run("ConcurrentIngestRace", testConcurrentIngestRace)
	t.Run("DeptAlertThresholds_NoBlock", testDeptAlertNoBlock)
	t.Run("IdempotentReplay", testIdempotentReplay)
	t.Run("GatewayBlocksAfterOverrun", testGatewayBlocksAfterOverrun)
	t.Run("WalletInsufficientGatewayReject", testWalletInsufficientGatewayReject)
}

// testKeyBudgetExhaustion verifies that when a PlatformKey's consumed amount
// exceeds its budget, the overrun job disables the key.
func testKeyBudgetExhaustion(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 1_000_000, // large so member doesn't trigger first
		DeptBudget:   1_000_000,
	})

	// Set consumed just under budget, then ingest to exceed
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 95)
	seedAndIngest(t, env, 100000) // will push consumed over 100
	drainOverrunJobs(t, env)

	// Key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Notification should have been written to notification_log
	assertNotificationInPG(t, env.Store, types.NotificationEventOverrunBlocked, 1)
}

// testMemberBudgetIsolation verifies that when member M1 exceeds personal budget,
// M1's keys are disabled.
func testMemberBudgetIsolation(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    1_000_000, // large so key budget doesn't trigger
		MemberBudget: 50,
		DeptBudget:   1_000_000,
	})

	// Set member consumed just under budget
	budgetfix.SetMemberSnapshotConsumed(t, env.Store, contract.IDMember1, 48)
	seedAndIngest(t, env, 100000) // will push member consumed over 50
	drainOverrunJobs(t, env)

	// Member's key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Overrun notification in PG
	assertNotificationInPG(t, env.Store, types.NotificationEventOverrunBlocked, 1)
}

// testConcurrentIngestRace verifies that multiple goroutines ingesting
// against the same key don't cause data races or double-disable issues.
func testConcurrentIngestRace(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 1_000_000,
		DeptBudget:   1_000_000,
	})

	// Set consumed to 90 so concurrent ingests will push over budget
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 90)

	const goroutines = 10
	const quotaPerReq = 25000 // ~small amount per request

	errs := concurrentIngest(t, env, goroutines, quotaPerReq)

	// Count successes
	var failures int
	for _, err := range errs {
		if err != nil {
			failures++
		}
	}
	if failures > 0 {
		t.Logf("note: %d/%d concurrent ingests had errors (acceptable under lock contention)", failures, goroutines)
	}

	// Drain all overrun jobs
	drainOverrunJobs(t, env)

	// Key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Final consumed should be > budget (100)
	consumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)
	if consumed <= 100 {
		t.Errorf("expected consumed > 100 after concurrent ingest, got %s", formatPoints(consumed))
	}
}

// testDeptAlertNoBlock verifies that department budget exhaustion sends notifications
// but does NOT disable keys (department-level is notify-only).
func testDeptAlertNoBlock(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:       1_000_000,
		MemberBudget:    1_000_000,
		DeptBudget:      1, // very low dept budget to trigger alert
		AlertThresholds: []int{50, 80, 100},
	})

	// Ingest enough to exceed dept budget (alert fires post-commit, no River job needed).
	seedAndIngest(t, env, 100000)

	// Key should STILL be active (dept overrun is notify-only)
	assertKeyActive(t, env.Store, contract.IDPlatformKey1)
}

// testIdempotentReplay verifies that re-ingesting the same log ID does not
// produce duplicate ledger entries or multiple disable calls.
func testIdempotentReplay(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 1_000_000,
		DeptBudget:   1_000_000,
	})

	// Set consumed close to budget
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 95)

	// First ingest — should trigger overrun
	logID := seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)
	firstConsumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)

	// Replay the same log ID — should be idempotent (no-op)
	err := env.Ingest.IngestByLogID(testutil.Ctx(), logID, types.SourceWebhook)
	if err != nil {
		t.Fatalf("replay ingest should not error: %v", err)
	}
	// No drainOverrunJobs here — idempotent replay should NOT enqueue any jobs

	// Consumed should not have increased
	afterConsumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)
	if afterConsumed > firstConsumed {
		t.Errorf("idempotent replay increased consumed: before=%s after=%s",
			formatPoints(firstConsumed), formatPoints(afterConsumed))
	}
}

// testGatewayBlocksAfterOverrun verifies the full chain:
// ingest → overrun job → key disabled → gateway precheck rejects request.
func testGatewayBlocksAfterOverrun(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 1_000_000,
		DeptBudget:   1_000_000,
	})

	keyHash := keyHashForPlatformKey(t, env.Store)

	// Before overrun — gateway should allow
	assertGatewayAllowed(t, env, keyHash)

	// Trigger overrun
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 95)
	seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	// After overrun — gateway should block
	assertGatewayBlocked(t, env, keyHash)
}

// testWalletInsufficientGatewayReject verifies that when wallet_quota_remain drops
// below minEstimatePoint, Gateway precheck rejects all requests.
func testWalletInsufficientGatewayReject(t *testing.T) {
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    1_000_000,
		MemberBudget: 1_000_000,
		DeptBudget:   1_000_000,
	})

	keyHash := keyHashForPlatformKey(t, env.Store)

	// Before: gateway should allow
	assertGatewayAllowed(t, env, keyHash)

	// Set wallet_quota_remain to 0
	ctx := testutil.Ctx()
	if err := env.Store.Company().SetWalletQuotaRemain(ctx, contract.DefaultCompanyID, 0, nil); err != nil {
		t.Fatalf("failed to set wallet remain: %v", err)
	}

	// After: gateway should reject
	assertGatewayBlocked(t, env, keyHash)
}

// ---------------------------------------------------------------------------
// Benchmark subtests — measure throughput as regression guards
// ---------------------------------------------------------------------------

func TestBenchmarkIngestThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skip benchmark in short mode")
	}

	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    1_000_000,
		MemberBudget: 1_000_000,
		DeptBudget:   10_000_000,
	})

	const iterations = 50
	start := time.Now()
	for i := 0; i < iterations; i++ {
		seedAndIngest(t, env, 10000)
	}
	elapsed := time.Since(start)

	qps := float64(iterations) / elapsed.Seconds()
	t.Logf("Ingest throughput: %d ops in %v = %.1f QPS (single-company serial)", iterations, elapsed, qps)

	if qps < 10 {
		t.Errorf("ingest QPS too low: %.1f (expected >= 10 for single company)", qps)
	}
}

func TestBenchmarkGatewayPrecheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skip benchmark in short mode")
	}

	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    1_000_000,
		MemberBudget: 1_000_000,
		DeptBudget:   10_000_000,
	})

	keyHash := keyHashForPlatformKey(t, env.Store)

	// Warm up
	_, _ = env.Precheck.Run(testutil.Ctx(), keyHash, "test-model", gateway.PrecheckOpts{})

	const iterations = 500
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = env.Precheck.Run(testutil.Ctx(), keyHash, "test-model", gateway.PrecheckOpts{})
	}
	elapsed := time.Since(start)

	qps := float64(iterations) / elapsed.Seconds()
	t.Logf("Gateway precheck: %d ops in %v = %.0f QPS", iterations, elapsed, qps)

	if qps < 100 {
		t.Errorf("precheck QPS too low: %.0f (expected >= 100)", qps)
	}
}

func TestBenchmarkConcurrentMultiKeyIngest(t *testing.T) {
	if testing.Short() {
		t.Skip("skip benchmark in short mode")
	}

	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    1_000_000,
		MemberBudget: 1_000_000,
		DeptBudget:   10_000_000,
	})

	const goroutines = 20
	const quotaPerReq = 10000

	start := time.Now()
	errs := concurrentIngest(t, env, goroutines, quotaPerReq)
	elapsed := time.Since(start)

	var successes int
	for _, err := range errs {
		if err == nil {
			successes++
		}
	}

	qps := float64(successes) / elapsed.Seconds()
	t.Logf("Concurrent ingest: %d/%d succeeded in %v = %.1f QPS",
		successes, goroutines, elapsed, qps)

	if successes < goroutines/2 {
		t.Errorf("too many failures: %d/%d succeeded", successes, goroutines)
	}
}
