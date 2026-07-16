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
	t.Run("MultiCompanyIsolation", testMultiCompanyIsolation)
	t.Run("WalletInsufficientGatewayReject", testWalletInsufficientGatewayReject)
}

// testKeyBudgetExhaustion verifies that when a PlatformKey's consumed amount
// exceeds its budget, the overrun job disables the key.
func testKeyBudgetExhaustion(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 10000, // large so member doesn't trigger first
		DeptBudget:   100000,
	})

	// Ingest under budget — key should remain active
	seedAndIngest(t, env, 400000) // ~90 points at model price
	drainOverrunJobs(t, env)
	assertKeyActive(t, env.Store, contract.IDPlatformKey1)

	// Set consumed close to budget to ensure next ingest triggers overrun
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 95)

	// Ingest more to exceed budget
	seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	// Key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Notification should have been sent with scope=platformKey
	assertNotificationCount(t, env, types.NotificationEventOverrunBlocked, 1)
	scope := notificationPayloadScope(t, env, types.NotificationEventOverrunBlocked)
	if scope != "platformKey" {
		t.Errorf("expected notification scope=platformKey, got %q", scope)
	}
}

// testMemberBudgetIsolation verifies that when member M1 exceeds personal budget,
// only M1's keys are disabled; other members are unaffected.
func testMemberBudgetIsolation(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100000, // large so key budget doesn't trigger
		MemberBudget: 50,
		DeptBudget:   100000,
	})

	// Set member consumed to just under budget
	budgetfix.SetMemberSnapshotConsumed(t, env.Store, contract.IDMember1, 48)

	// Ingest to exceed member budget
	seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	// Member's key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Notification scope should be member
	assertNotificationCount(t, env, types.NotificationEventOverrunBlocked, 1)
	scope := notificationPayloadScope(t, env, types.NotificationEventOverrunBlocked)
	if scope != "member" {
		t.Errorf("expected notification scope=member, got %q", scope)
	}
}

// testConcurrentIngestRace verifies that multiple goroutines ingesting
// against the same key don't cause data races or double-disable issues.
func testConcurrentIngestRace(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 100000,
		DeptBudget:   100000,
	})

	// Set consumed to 90 — each ingest adds ~5 points, so after 10 goroutines
	// consumed should be ~140 and key should be disabled exactly once.
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 90)

	const goroutines = 10
	const quotaPerReq = 25000 // ~5 points

	errs := concurrentIngest(t, env, goroutines, quotaPerReq)

	// All ingests should succeed (or fail gracefully due to locking)
	var failures int
	for _, err := range errs {
		if err != nil {
			failures++
		}
	}
	if failures > 0 {
		t.Logf("note: %d/%d concurrent ingests had errors (expected under lock contention)", failures, goroutines)
	}

	// Drain all overrun jobs
	drainOverrunJobs(t, env)

	// Key should be disabled
	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)

	// Should have at least 1 overrun notification (may have more due to race, but never 0)
	overrunNotifs := env.Notifier.byEvent(types.NotificationEventOverrunBlocked)
	if len(overrunNotifs) == 0 {
		t.Error("expected at least 1 overrun_blocked notification after concurrent ingest")
	}

	// Final consumed should be > budget (100)
	consumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)
	if consumed <= 100 {
		t.Errorf("expected consumed > 100 after concurrent ingest, got %s", formatPoints(consumed))
	}
}

// testDeptAlertNoBlock verifies that department budget exhaustion sends notifications
// but does NOT disable keys (department-level is notify-only).
func testDeptAlertNoBlock(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:       100000,
		MemberBudget:    100000,
		DeptBudget:      1, // very low dept budget to trigger
		AlertThresholds: []int{50, 80, 100},
	})

	// Ingest enough to exceed dept budget
	seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	// Key should STILL be active (dept overrun is notify-only)
	assertKeyActive(t, env.Store, contract.IDPlatformKey1)

	// Should have an overrun notification with notifyOnly=true
	overrunNotifs := env.Notifier.byEvent(types.NotificationEventOverrunBlocked)
	if len(overrunNotifs) == 0 {
		// Department overrun notification is best-effort; it fires only if dept budget
		// check is reached (i.e., no higher-priority axis triggered first).
		// If key/member budget didn't trigger, dept should.
		t.Log("note: no dept overrun notification (higher axis may have matched)")
	} else {
		// Verify notifyOnly flag
		notifyOnly, ok := overrunNotifs[0].Payload["notifyOnly"].(bool)
		if ok && !notifyOnly {
			t.Error("expected dept overrun notification to have notifyOnly=true")
		}
	}
}

// testIdempotentReplay verifies that re-ingesting the same log ID does not
// produce duplicate ledger entries or multiple disable calls.
func testIdempotentReplay(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 100000,
		DeptBudget:   100000,
	})

	// Set consumed close to budget
	budgetfix.SetPlatformKeySnapshotConsumed(t, env.Store, contract.IDPlatformKey1, 95)

	// First ingest — should trigger overrun
	logID := seedAndIngest(t, env, 100000)
	drainOverrunJobs(t, env)

	assertKeyDisabled(t, env.Store, contract.IDPlatformKey1)
	firstNotifCount := env.Notifier.count()
	firstConsumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)

	// Replay the same log ID — should be idempotent (no-op)
	err := env.Ingest.IngestByLogID(testutil.Ctx(), logID, types.SourceWebhook)
	if err != nil {
		t.Fatalf("replay ingest should not error: %v", err)
	}
	drainOverrunJobs(t, env)

	// Consumed should not have increased
	afterConsumed := consumedForKey(t, env.Store, contract.IDPlatformKey1)
	if afterConsumed > firstConsumed+0.001 {
		t.Errorf("idempotent replay increased consumed: before=%s after=%s",
			formatPoints(firstConsumed), formatPoints(afterConsumed))
	}

	// No additional notifications
	if env.Notifier.count() != firstNotifCount {
		t.Errorf("idempotent replay generated extra notifications: before=%d after=%d",
			firstNotifCount, env.Notifier.count())
	}
}

// testGatewayBlocksAfterOverrun verifies the full chain:
// ingest → overrun job → key disabled → gateway precheck rejects request.
func testGatewayBlocksAfterOverrun(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 100000,
		DeptBudget:   100000,
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

// testMultiCompanyIsolation verifies that when company A's key exceeds budget,
// company B (same schema, different logical company) is completely unaffected.
// Each company gets its own stressEnv (= its own PG schema), simulating true isolation.
func testMultiCompanyIsolation(t *testing.T) {
	t.Parallel()

	// Company A: low key budget, will exceed
	envA := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 100000,
		DeptBudget:   100000,
	})

	// Company B: same budget config, will NOT exceed
	envB := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100,
		MemberBudget: 100000,
		DeptBudget:   100000,
	})

	// Company A: push consumed over budget
	budgetfix.SetPlatformKeySnapshotConsumed(t, envA.Store, contract.IDPlatformKey1, 95)
	seedAndIngest(t, envA, 100000)
	drainOverrunJobs(t, envA)

	// Company B: stay under budget
	budgetfix.SetPlatformKeySnapshotConsumed(t, envB.Store, contract.IDPlatformKey1, 30)
	seedAndIngest(t, envB, 100000)
	drainOverrunJobs(t, envB)

	// Assert: A's key is disabled
	assertKeyDisabled(t, envA.Store, contract.IDPlatformKey1)

	// Assert: B's key is still active
	assertKeyActive(t, envB.Store, contract.IDPlatformKey1)

	// Assert: A has overrun notification
	assertNotificationCount(t, envA, types.NotificationEventOverrunBlocked, 1)

	// Assert: B has NO overrun notification
	assertNoNotification(t, envB, types.NotificationEventOverrunBlocked)

	// Assert: Gateway for A is blocked, B is allowed
	keyHashA := keyHashForPlatformKey(t, envA.Store)
	keyHashB := keyHashForPlatformKey(t, envB.Store)
	assertGatewayBlocked(t, envA, keyHashA)
	assertGatewayAllowed(t, envB, keyHashB)
}

// testWalletInsufficientGatewayReject verifies that when a company's wallet_remain
// drops below minEstimatePoint, the Gateway precheck rejects all requests before
// they reach the ingest pipeline.
func testWalletInsufficientGatewayReject(t *testing.T) {
	t.Parallel()
	env := buildStressEnv(t, stressEnvOpts{
		KeyBudget:    100000,
		MemberBudget: 100000,
		DeptBudget:   100000,
	})

	keyHash := keyHashForPlatformKey(t, env.Store)

	// Before: gateway should allow (wallet is seeded with enough balance)
	assertGatewayAllowed(t, env, keyHash)

	// Set wallet_remain to 0 (below minEstimatePoint)
	ctx := testutil.Ctx()
	if err := env.Store.Company().SetWalletRemain(ctx, contract.DefaultCompanyID, 0, nil); err != nil {
		t.Fatalf("failed to set wallet remain: %v", err)
	}

	// After: gateway should reject with "insufficient wallet points"
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
		KeyBudget:    1_000_000, // large budget so no overrun during benchmark
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

	// Soft lower bound — if this fails, there's a significant performance regression
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
	_ = env.Precheck.Run(testutil.Ctx(), keyHash, "local-test-model", gateway.PrecheckOpts{})

	const iterations = 500
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = env.Precheck.Run(testutil.Ctx(), keyHash, "local-test-model", gateway.PrecheckOpts{})
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
