package usage_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func TestIngestIdempotentAndRollup(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	beforeUsed := testutil.PlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1)
	beforeConsumed := testutil.Dept3SnapshotConsumed(t, st)
	beforeBuckets := testutil.UsageBucketCount(st)

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(1001, 99))
	if err := ingest.IngestByLogID(testutil.Ctx(), 1001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if err := ingest.IngestByLogID(testutil.Ctx(), 1001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	exists, err := testutil.HasLedgerLogID(st, 1001)
	if err != nil || !exists {
		t.Fatalf("expected ledger entry for log 1001, exists=%v err=%v", exists, err)
	}

	afterUsed := testutil.PlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1)
	if afterUsed <= beforeUsed {
		t.Fatalf("expected key used increase, before=%v after=%v", beforeUsed, afterUsed)
	}

	afterConsumed := testutil.Dept3SnapshotConsumed(t, st)
	if afterConsumed <= beforeConsumed {
		t.Fatalf("expected consumed rollup, before=%v after=%v", beforeConsumed, afterConsumed)
	}

	afterBuckets := testutil.UsageBucketCount(st)
	if afterBuckets < beforeBuckets+1 {
		t.Fatalf("expected at least one new usage bucket, before=%d after=%d", beforeBuckets, afterBuckets)
	}
	if err := ingest.IngestByLogID(testutil.Ctx(), 1001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	testutil.AssertUsageBucketCount(t, st, afterBuckets)
}

func TestIngestByLogID(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(2002, 99))
	if err := ingest.IngestByLogID(testutil.Ctx(), 2002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(st, 2002)
	if err != nil || !ingested {
		t.Fatalf("expected log 2002 in ledger, err=%v ingested=%v", err, ingested)
	}
}

func TestIngestWritesUsageBucket(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
	beforeBuckets := testutil.UsageBucketCount(st)

	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{
		ID: 4001, TokenID: 99, Quota: 100000, ModelName: "gpt-4o", CreatedAt: 1717200000,
	})
	if err := ingest.IngestByLogID(testutil.Ctx(), 4001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	afterBuckets := testutil.UsageBucketCount(st)
	if afterBuckets < beforeBuckets+1 {
		t.Fatalf("expected at least one new usage bucket, before=%d after=%d", beforeBuckets, afterBuckets)
	}

	if err := ingest.IngestByLogID(testutil.Ctx(), 4001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	testutil.AssertUsageBucketCount(t, st, afterBuckets)
}

func TestIngestRaw(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	raw := testutil.DefaultConsumeLog(3003, 99)
	if err := ingest.IngestRaw(testutil.Ctx(), raw, types.SourceReconcile); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(st, 3003)
	if err != nil || !ingested {
		t.Fatalf("expected log 3003 in ledger, err=%v ingested=%v", err, ingested)
	}
}

func TestIngestStoresLedgerPeriodKey(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
	raw := testutil.DefaultConsumeLog(8801, 99)
	testutil.SeedConsumeLog(t, st, raw)
	if err := ingest.IngestByLogID(testutil.Ctx(), 8801, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	want := contract.DemoBudgetPeriod
	entries, _, err := st.Ledger().ListCallSettledPage(testutil.Ctx(), store.LedgerCallFilter{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	var found *types.UsageLedgerEntry
	for i := range entries {
		if entries[i].IdempotencyKey == usage.NewAPIIdempotencyKey(8801) {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected ledger entry for log 8801")
	}
	if found.PeriodKey != want {
		t.Fatalf("PeriodKey = %q, want %q", found.PeriodKey, want)
	}
}

func TestIngestSnapshotUsesNowPeriodForMonthlyOrg(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ctx := testutil.Ctx()
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !setOrgNodePeriodMonthly(nodes, contract.IDDept3) {
		t.Fatal("dept-3 not found")
	}
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}

	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	occurred := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	snapshotPeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	ledgerPeriod := pkgbudget.OccurrenceSnapshotKey(pkgbudget.PeriodMonthly, occurred).String()
	if snapshotPeriod == ledgerPeriod {
		t.Skip("requires occurred month to differ from current month")
	}

	raw := testutil.DefaultConsumeLog(9901, 99)
	raw.CreatedAt = occurred.Unix()
	testutil.SeedConsumeLog(t, st, raw)

	beforeSnapshot := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, snapshotPeriod)
	beforeLedgerPeriod := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, ledgerPeriod)

	if err := ingest.IngestByLogID(ctx, 9901, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	afterSnapshot := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, snapshotPeriod)
	if afterSnapshot <= beforeSnapshot {
		t.Fatalf("expected snapshot period %q consumption increase, before=%v after=%v", snapshotPeriod, beforeSnapshot, afterSnapshot)
	}
	afterLedgerPeriod := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, ledgerPeriod)
	if afterLedgerPeriod != beforeLedgerPeriod {
		t.Fatalf("expected no consumption at ledger period %q, before=%v after=%v", ledgerPeriod, beforeLedgerPeriod, afterLedgerPeriod)
	}
}

func setOrgNodePeriodMonthly(nodes []types.OrgNode, id string) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Period = pkgbudget.PeriodMonthly
			return true
		}
		if len(nodes[i].Children) > 0 && setOrgNodePeriodMonthly(nodes[i].Children, id) {
			return true
		}
	}
	return false
}

func TestIngestByLogIDNotFound(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)

	err := ingest.IngestByLogID(testutil.Ctx(), 9999, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected error for missing consume log")
	}
	if !errors.Is(err, store.ErrConsumeLogNotFound) {
		t.Fatalf("expected ErrConsumeLogNotFound, got %v", err)
	}
}

func TestIngestByLogIDMappingMissing(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(5005, 42))

	err := ingest.IngestByLogID(testutil.Ctx(), 5005, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected mapping missing error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusNotFound {
		t.Fatalf("expected not found domain error, got %v", err)
	}
}
