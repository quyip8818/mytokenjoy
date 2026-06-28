package budget_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestIdempotentAndRollup(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ingest := budget.NewIngestService(cfg, st, lifecycle, logger)

	mapping := store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: testutil.Int64Ptr(99),
		MemberID:      testutil.StrPtr(seed.IDMember1),
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}
	if err := st.Relay().UpsertMapping(mapping); err != nil {
		t.Fatal(err)
	}

	keys := st.Keys().PlatformKeys()
	for i := range keys {
		if keys[i].ID == seed.IDPlatformKey1 {
			keys[i].Used = 0
			st.Keys().SetPlatformKeys(keys)
			break
		}
	}

	tree := st.Budget().Tree()
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
	if err := ingest.Ingest(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if err := ingest.Ingest(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	exists, err := st.Relay().HasIngestedLogID(1001)
	if err != nil || !exists {
		t.Fatalf("expected ingested log id, exists=%v err=%v", exists, err)
	}

	keys = st.Keys().PlatformKeys()
	for _, key := range keys {
		if key.ID == seed.IDPlatformKey1 && key.Used <= 0 {
			t.Fatalf("expected key used > 0, got %v", key.Used)
		}
	}

	tree = st.Budget().Tree()
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
}

func TestIngestFromOutbox(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ingest := budget.NewIngestService(cfg, st, lifecycle, logger)

	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: testutil.Int64Ptr(99),
		MemberID:      testutil.StrPtr(seed.IDMember1),
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(newapi.WebhookLogPayload{
		ID: 2002, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ingest.IngestFromOutbox(context.Background(), raw); err != nil {
		t.Fatal(err)
	}
	ingested, err := st.Relay().HasIngestedLogID(2002)
	if err != nil || !ingested {
		t.Fatalf("expected log 2002 ingested via outbox, err=%v ingested=%v", err, ingested)
	}
}

func TestEnqueueFailed(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ingest := budget.NewIngestService(cfg, st, lifecycle, logger)

	payload := newapi.WebhookLogPayload{ID: 3003, TokenID: 1, Quota: 100, Model: "gpt-4o"}
	if err := ingest.EnqueueFailed(payload, errors.New("simulated failure")); err != nil {
		t.Fatal(err)
	}
	entries, err := st.Relay().ClaimPendingWebhookOutbox(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected pending webhook outbox entry after EnqueueFailed")
	}
}
