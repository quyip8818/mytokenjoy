package worker_test

import (
	"encoding/json"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestWorkerProcessesRebalanceQueue(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	tokenID := int64(42)
	remainQuota := int64(1000)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		CompanyID: seed.DefaultCompanyID, PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID, MemberID: testutil.StrPtr(seed.IDMember1),
		DepartmentID: seed.IDDept3, SyncStatus: store.RelaySyncStatusSynced,
		RelayGroup: "dept-dept-3", NewAPITokenRemainQuota: &remainQuota,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisMember, seed.IDMember1); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected rebalance processor to update token")
	}
	if testutil.PendingRebalanceCount(st, seed.DefaultCompanyID) != 0 {
		t.Fatal("expected rebalance queue drained")
	}
}

func TestWorkerProcessesOverrunQueue(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, _ := newWorkerRunner(t, stub)
	ctx := testutil.Ctx()

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 25000)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, err := json.Marshal(map[string]string{
		"departmentId": seed.IDDept3, "platformKeyId": seed.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Relay().EnqueueOverrun(ctx, payload); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == seed.IDPlatformKey1 && key.Status == "active" {
			t.Fatalf("expected plk-1 disabled after overrun processor, status=%q", key.Status)
		}
	}
}
