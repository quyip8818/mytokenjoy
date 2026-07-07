package usage_test

import (
	"errors"
	"testing"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestIdempotentAndRollup(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
	ctx := testutil.Ctx()

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == seed.IDPlatformKey1 {
			keys[i].Used = 0
			if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
				t.Fatal(err)
			}
			break
		}
	}

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	var leaf *types.BudgetNode
	var walk func([]types.BudgetNode)
	walk = func(nodes []types.BudgetNode) {
		for i := range nodes {
			if nodes[i].ID == seed.IDDept3 {
				leaf = &nodes[i]
			}
			if len(nodes[i].Children) > 0 {
				walk(nodes[i].Children)
			}
		}
	}
	walk(tree)
	if leaf == nil {
		t.Fatal("dept-3 not found")
	}
	beforeConsumed := leaf.Consumed

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

	keys, err = st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == seed.IDPlatformKey1 && key.Used <= 0 {
			t.Fatalf("expected key used > 0, got %v", key.Used)
		}
	}

	tree, err = common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	walk = func(nodes []types.BudgetNode) {
		for i := range nodes {
			if nodes[i].ID == seed.IDDept3 {
				leaf = &nodes[i]
			}
			if len(nodes[i].Children) > 0 {
				walk(nodes[i].Children)
			}
		}
	}
	walk(tree)
	if leaf.Consumed <= beforeConsumed {
		t.Fatalf("expected consumed rollup, before=%v after=%v", beforeConsumed, leaf.Consumed)
	}

	testutil.AssertUsageBucketCount(t, st, 1)
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

	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{
		ID: 4001, TokenID: 99, Quota: 100000, ModelName: "gpt-4o", CreatedAt: 1717200000,
	})
	if err := ingest.IngestByLogID(testutil.Ctx(), 4001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	testutil.AssertUsageBucketCount(t, st, 1)

	if err := ingest.IngestByLogID(testutil.Ctx(), 4001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	testutil.AssertUsageBucketCount(t, st, 1)
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
