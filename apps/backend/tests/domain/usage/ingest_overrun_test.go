package usage_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
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
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	ingest := usage.NewIngestService(cfg, st, notifier, logger)
	overrun := budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	rebalance := budget.NewRebalanceService(cfg, st, stub)
	orgSvc := testutil.NewOrgService(t, cfg, st)
	runner := worker.NewRunner(cfg, st.Relay(), stub, lifecycle, ingest, overrun, rebalance, orgSvc, logger)
	ctx := testutil.Ctx()

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 24999)
	testutil.PersistBudgetTreeT(t, ctx, st, tree)

	testutil.UpsertRelayMapping(t, st, testutil.DefaultRelayMappingOpts())

	payload := newapi.WebhookLogPayload{
		ID: 3001, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	}
	if err := ingest.Ingest(testutil.Ctx(), payload, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(ctx)

	tree, err = common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := findDept3(tree)
	if node == nil || node.Consumed < node.Budget {
		t.Fatalf("expected dept-3 consumed >= budget, consumed=%v budget=%v", node.Consumed, node.Budget)
	}

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var plk1 *types.PlatformKey
	for _, key := range keys {
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
