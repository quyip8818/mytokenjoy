package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestIngestOverrunDisablesDepartmentKeys(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, _, ingest := workerfix.NewRunner(t, stub)
	ctx := testutil.Ctx()

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 24999)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)

	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(3001, 99))
	if err := ingest.IngestByLogID(testutil.Ctx(), 3001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(ctx)

	tree, err = common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := budget.FindBudgetNode(tree, seed.IDDept3)
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
