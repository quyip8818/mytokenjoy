package worker_test

import (
	"encoding/json"
	"testing"

	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"

	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestWorkerProcessesRebalanceQueue(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	tokenID := int64(42)
	remainQuota := int64(1000)
	if err := fix.st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID: contract.DefaultCompanyID, PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID: &tokenID, MemberID: testutil.StrPtr(contract.IDMember1),
		DepartmentID: contract.IDDept3, SyncStatus: store.MappingSyncStatusSynced,
		NewAPIGroup: "dept-dept-3", NewAPIKeyRemainQuota: &remainQuota,
	}); err != nil {
		t.Fatal(err)
	}
	if err := jobs.InsertRebalance(ctx, fix.rt.Enqueuer, nil, contract.DefaultCompanyID, store.RebalanceAxisMember, contract.IDMember1); err != nil {
		t.Fatal(err)
	}

	fix.runRiver(t)
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected rebalance processor to update token")
	}
	if testutil.PendingRebalanceCount(fix.st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected rebalance queue drained")
	}
}

func TestWorkerProcessesOverrunQueue(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, fix.st, newapisynctf.DefaultMappingOpts())
	testutil.SetDeptSnapshotConsumed(t, fix.st, contract.IDDept3, testutil.DisplayPoints(25000))

	payload, err := json.Marshal(map[string]string{
		"departmentId": contract.IDDept3, "platformKeyId": contract.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := jobs.InsertOverrun(ctx, fix.rt.Enqueuer, nil, contract.DefaultCompanyID, payload); err != nil {
		t.Fatal(err)
	}

	fix.runRiver(t)

	keys, err := fix.st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 && key.Status == "active" {
			t.Fatalf("expected plk-1 disabled after overrun processor, status=%q", key.Status)
		}
	}
}
