package budget_test

import (
	"context"
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
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestIngestOverrunDisablesDepartmentKeys(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
		testutil.WithNewAPIWebhookSecret("secret"),
	)
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ingest := budget.NewIngestService(cfg, st, lifecycle, logger)

	tree := st.Budget().Tree()
	setDept3Consumed(tree, 24999)
	st.Budget().SetTree(tree)

	tokenID := int64(99)
	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      testutil.StrPtr(seed.IDMember1),
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	payload := newapi.WebhookLogPayload{
		ID: 3001, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	}
	if err := ingest.Ingest(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	node := findDept3(st.Budget().Tree())
	if node == nil || node.Consumed < node.Budget {
		t.Fatalf("expected dept-3 consumed >= budget, consumed=%v budget=%v", node.Consumed, node.Budget)
	}

	var plk1 *types.PlatformKey
	for _, key := range st.Keys().PlatformKeys() {
		if key.ID == seed.IDPlatformKey1 {
			copy := key
			plk1 = &copy
			break
		}
	}
	if plk1 == nil {
		t.Fatal("plk-1 not found")
	}
	if plk1.Status == "active" {
		t.Fatalf("expected plk-1 disabled after department overrun, status=%q", plk1.Status)
	}
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken call when disabling relay token")
	}
}

func setDept3Consumed(tree []types.BudgetNode, consumed float64) {
	var walk func([]*types.BudgetNode)
	walk = func(nodes []*types.BudgetNode) {
		for i := range nodes {
			if nodes[i].ID == seed.IDDept3 {
				nodes[i].Consumed = consumed
				return
			}
			if len(nodes[i].Children) > 0 {
				children := make([]*types.BudgetNode, len(nodes[i].Children))
				for j := range nodes[i].Children {
					children[j] = &nodes[i].Children[j]
				}
				walk(children)
			}
		}
	}
	roots := make([]*types.BudgetNode, len(tree))
	for i := range tree {
		roots[i] = &tree[i]
	}
	walk(roots)
}
