package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestIngestOverrunDisablesDepartmentKeys(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, _, ingest := workerfix.NewRunner(t, stub)
	ctx := testutil.Ctx()

	testutil.SetDeptSnapshotConsumed(t, st, contract.IDDept3, 0)

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(3001, 99))
	if err := ingest.IngestByLogID(testutil.Ctx(), 3001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	testutil.SetDeptSnapshotConsumed(t, st, contract.IDDept3, testutil.DisplayPoints(25000))
	runner.RunOnce(ctx)

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	consumed := testutil.SnapshotConsumed(t, st, store.SnapshotAxisOrgNode, contract.IDDept3)
	if node == nil || consumed < node.Budget {
		t.Fatalf("expected dept-3 consumed >= budget, consumed=%v budget=%v", consumed, node.Budget)
	}

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
	if plk1.Status == "active" {
		t.Fatalf("expected plk-1 disabled after department overrun, status=%q", plk1.Status)
	}
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken call when disabling platform key via newapi")
	}
}
