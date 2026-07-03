package budget_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestIdempotentAndRollup(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	testutil.UpsertRelayMapping(t, st, testutil.DefaultRelayMappingOpts())
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

	tree, err := common.LoadBudgetTree(ctx, st)
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

	payload := newapi.WebhookLogPayload{
		ID: 1001, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	}
	if err := ingest.Ingest(testutil.Ctx(), payload, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if err := ingest.Ingest(testutil.Ctx(), payload, types.SourceWebhook); err != nil {
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

	tree, err = common.LoadBudgetTree(ctx, st)
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

func TestIngestFromOutbox(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	testutil.UpsertRelayMapping(t, st, testutil.DefaultRelayMappingOpts())

	raw, err := json.Marshal(newapi.WebhookLogPayload{
		ID: 2002, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ingest.IngestFromOutbox(testutil.Ctx(), raw); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(st, 2002)
	if err != nil || !ingested {
		t.Fatalf("expected log 2002 in ledger via outbox, err=%v ingested=%v", err, ingested)
	}
}

func TestIngestWritesUsageBucket(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	testutil.UpsertRelayMapping(t, st, testutil.DefaultRelayMappingOpts())

	payload := newapi.WebhookLogPayload{
		ID: 4001, TokenID: 99, Quota: 100000, Model: "gpt-4o", CreatedAt: 1717200000,
	}
	if err := ingest.Ingest(testutil.Ctx(), payload, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	testutil.AssertUsageBucketCount(t, st, 1)

	if err := ingest.Ingest(testutil.Ctx(), payload, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	testutil.AssertUsageBucketCount(t, st, 1)
}

func TestEnqueueFailed(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	ctx := testutil.Ctx()

	payload := newapi.WebhookLogPayload{ID: 3003, TokenID: 1, Quota: 100, Model: "gpt-4o"}
	if err := ingest.EnqueueFailed(ctx, payload, errors.New("simulated failure")); err != nil {
		t.Fatal(err)
	}
	entries, err := st.Relay().ClaimPendingWebhookOutbox(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected pending webhook outbox entry after EnqueueFailed")
	}
}
