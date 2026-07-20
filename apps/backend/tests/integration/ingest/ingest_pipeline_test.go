//go:build testhook && integration

package ingest_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

// TestIngestIdempotentAndRollup verifies the full ingest→River→budget pipeline:
// - Idempotent: double-ingest produces a single ledger entry
// - Consumed snapshots increase for both key and member axes
func TestIngestIdempotentAndRollup(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	beforeKeyConsumed := budgetfix.PlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1)
	beforeMemberConsumed := budgetfix.SnapshotConsumed(t, st, store.AxisKindMember, contract.IDMember1)

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(1001, 99))
	if err := ingest.IngestByLogID(ctx, 1001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if err := ingest.IngestByLogID(ctx, 1001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(t, ctx)

	exists, err := testutil.HasLedgerLogID(st, 1001)
	if err != nil || !exists {
		t.Fatalf("expected ledger entry for log 1001, exists=%v err=%v", exists, err)
	}

	afterKeyConsumed := budgetfix.PlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1)
	if afterKeyConsumed <= beforeKeyConsumed {
		t.Fatalf("expected key consumed increase, before=%v after=%v", beforeKeyConsumed, afterKeyConsumed)
	}

	afterMemberConsumed := budgetfix.SnapshotConsumed(t, st, store.AxisKindMember, contract.IDMember1)
	if afterMemberConsumed <= beforeMemberConsumed {
		t.Fatalf("expected member consumed increase, before=%v after=%v", beforeMemberConsumed, afterMemberConsumed)
	}
}

// TestIngestSnapshotUsesNowPeriodForMonthlyOrg verifies that budget_consumed
// is written to the *current* budget period (open period) even when the ingest
// event occurred in a past month.
func TestIngestSnapshotUsesNowPeriodForMonthlyOrg(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()
	cfg := runner.Cfg
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !orgfix.SetNodePeriod(nodes, contract.IDDept3, pkgbudget.PeriodMonthly) {
		t.Fatal("dept-3 not found")
	}
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	occurred := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	snapshotPeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	ledgerPeriod := pkgbudget.OccurrenceSnapshotKey(pkgbudget.PeriodMonthly, occurred).String()
	if snapshotPeriod == ledgerPeriod {
		t.Skip("requires occurred month to differ from current month")
	}

	raw := testutil.DefaultConsumeLog(9901, 99)
	raw.CreatedAt = occurred.Unix()
	testutil.SeedConsumeLog(t, st, raw)

	beforeSnapshot := budgetfix.SnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, snapshotPeriod)
	beforeLedgerPeriod := budgetfix.SnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, ledgerPeriod)

	if err := ingest.IngestByLogID(ctx, 9901, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(t, ctx)

	afterSnapshot := budgetfix.SnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, snapshotPeriod)
	if afterSnapshot <= beforeSnapshot {
		t.Fatalf("expected snapshot period %q consumption increase, before=%v after=%v", snapshotPeriod, beforeSnapshot, afterSnapshot)
	}
	afterLedgerPeriod := budgetfix.SnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, ledgerPeriod)
	if afterLedgerPeriod != beforeLedgerPeriod {
		t.Fatalf("expected no consumption at ledger period %q, before=%v after=%v", ledgerPeriod, beforeLedgerPeriod, afterLedgerPeriod)
	}
}

// TestIngestAppKeyIncrementsPlatformKeyConsumed verifies that app-scoped keys
// (project keys without a member) still have their budget_consumed incremented.
func TestIngestAppKeyIncrementsPlatformKeyConsumed(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 77, RemainQuota: 1000}}
	_, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()

	fullKey := "sk-app-key-test"
	plk3 := uuid.MustParse("00000000-0000-7000-8000-000000000f13")
	if err := st.Keys().SetPlatformKeys(ctx, []types.PlatformKey{{
		ID:        plk3,
		Name:      "App Key",
		KeyPrefix: "sk-app",
		Scope:     types.PlatformKeyScopeProject,
		FullKey:   &fullKey,
		Status:    "active",
		CreatedAt: "2026-06-19",
	}}); err != nil {
		t.Fatal(err)
	}

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.MappingOpts{
		PlatformKeyID: plk3,
		NewAPIKeyID:   77,
		NoMember:      true,
		DepartmentID:  contract.IDDept3,
	})

	before := budgetfix.PlatformKeySnapshotConsumed(t, st, plk3)

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(98002, 77))
	if err := ingest.IngestByLogID(ctx, 98002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	after := budgetfix.PlatformKeySnapshotConsumed(t, st, plk3)
	if after <= before {
		t.Fatalf("expected platform key consumed increase for app key, before=%v after=%v", before, after)
	}
}

// TestIngestDoesNotEnqueueDashboard verifies that after a successful ingest,
// dashboard_project is NOT enqueued (driven by hourly watchdog instead).
func TestIngestDoesNotEnqueueDashboard(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	_, st, ingest := riverfix.NewIngestRuntime(t, stub)

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	// Seed combined_key_remain so overrun safety-enqueue doesn't trigger.
	budgetfix.SetCombinedKeyRemain(t, st, contract.IDPlatformKey1, 999_999_999)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4101, 99))

	if err := ingest.IngestByLogID(testutil.Ctx(), 4101, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	if riverfix.PendingDashboardProjectCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no dashboard_project job from ingest (now driven by watchdog)")
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no rebalance jobs directly from ingest")
	}
	if riverfix.PendingOverrunCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no overrun jobs directly from ingest when remain is high")
	}
}
