package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestIngestOverrunNotifiesDepartmentWithoutDisablingKeys(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !orgfix.SetNodeBudget(nodes, contract.IDDept3, 1) {
		t.Fatal("dept-3 not found")
	}
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(types.OrgNodesToBudgetTree(nodes), contract.IDDept3)
	if node == nil {
		t.Fatal("budget node not found")
	}
	node.Budget = 1
	if err := pkgbudget.PersistNodeBudget(ctx, st.Budget().OrgNodeBudget(), contract.IDDept3, *node); err != nil {
		t.Fatal(err)
	}

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(3001, 99))
	if err := ingest.IngestByLogID(testutil.Ctx(), 3001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(t, ctx)

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var plk1 *types.PlatformKey
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			copy := key
			plk1 = &copy
			break
		}
	}
	if plk1 == nil {
		t.Fatal("plk-1 not found")
	}
	if plk1.Status != "active" {
		t.Fatalf("expected plk-1 to remain active after department ledger overrun, status=%q", plk1.Status)
	}
}
