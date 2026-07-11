package keys_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApprovalBudgetCheckInsufficient(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	check, err := svc.ApprovalBudgetCheck(testutil.Ctx(), contract.IDApproval1)
	if err != nil {
		t.Fatal(err)
	}
	if check.Sufficient {
		t.Fatal("expected apv-1 insufficient (dept-4 has no reserved pool)")
	}
	if check.ReservedPool != 0 {
		t.Fatalf("expected reserved pool 0, got %v", check.ReservedPool)
	}
}

func TestApprovalBudgetCheckSufficient(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "budget", Reason: "test", RequestedBudget: 1000,
		RequestedModels: []int64{contract.IDModel1}, MemberID: contract.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	check, err := svc.ApprovalBudgetCheck(testutil.Ctx(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !check.Sufficient {
		t.Fatalf("expected sufficient, reserved=%v requested=%v", check.ReservedPool, check.Requested)
	}
	_ = st
}

func TestApproveKeyTypeCreatesPlatformKey(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	keysBefore, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before := len(keysBefore)
	if err := svc.ApproveApproval(testutil.Ctx(), contract.IDApproval1, contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	approval := findApproval(st, contract.IDApproval1)
	if approval == nil || approval.Status != "approved" {
		t.Fatalf("expected apv-1 approved, got %+v", approval)
	}
	after, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != before+1 {
		t.Fatalf("expected one new platform key, before=%d after=%d", before, len(after))
	}
}

func TestApproveQuotaTypeAddsPersonalBudget(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "budget", Reason: "need more", RequestedBudget: 1000,
		RequestedModels: []int64{contract.IDModel1}, MemberID: contract.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	membersBefore, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	before := budget.GetPersonalBudget(membersBefore, contract.IDMember1)
	if err := svc.ApproveApproval(testutil.Ctx(), created.ID, contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	membersAfter, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	after := budget.GetPersonalBudget(membersAfter, contract.IDMember1)
	if after != before+1000 {
		t.Fatalf("expected personal quota +1000, before=%v after=%v", before, after)
	}
}

func TestApproveInsufficientReserved(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "budget", Reason: "too much", RequestedBudget: 2_000_000,
		RequestedModels: []int64{contract.IDModel1}, MemberID: contract.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = svc.ApproveApproval(testutil.Ctx(), created.ID, contract.IDMemberAdmin)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestRejectApproval(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	reason := "not needed"
	if err := svc.RejectApproval(testutil.Ctx(), contract.IDApproval2, contract.IDMemberAdmin, &reason); err != nil {
		t.Fatal(err)
	}
	approval := findApproval(st, contract.IDApproval2)
	if approval == nil || approval.Status != "rejected" {
		t.Fatalf("expected apv-2 rejected, got %+v", approval)
	}
}

func TestCreatePlatformKeyRequiresNewAPI(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "test-key", MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
}

func TestCreatePlatformKeySuccess(t *testing.T) {
	t.Parallel()
	svc, _, _ := newKeysServiceWithNewAPI(t)
	memberID := contract.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "test-key", MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.FullKey == nil || *created.FullKey != "sk-test-key" {
		t.Fatalf("expected platform key full key, got %+v", created.FullKey)
	}
}

func TestCreatePlatformKeyQuotaExceeded(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "too-big", MemberID: &memberID, Budget: 20_000_000,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreatePlatformKeyInvalidWhitelist(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "bad-models", MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []int64{999999},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateApprovalInvalidModels(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	_, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "budget", Reason: "bad models", RequestedBudget: 1000,
		RequestedModels: []int64{999999}, MemberID: contract.IDMember1,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateGroupKeyQuotaExceeded(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	groupID := contract.IDBudgetGroup1
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "group-over", BudgetGroupID: &groupID, MemberID: &memberID, Budget: 20_000_000,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdatePlatformKeyQuota(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	quota := 20_000_000.0
	_, err := svc.UpdatePlatformKey(testutil.Ctx(), contract.IDPlatformKey1, types.UpdatePlatformKeyInput{
		Budget: &quota,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestDeletePlatformKeyReleasesQuota(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	memberID := contract.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "release-me", MemberID: &memberID, Budget: 500,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeSummary, err := svc.BudgetSummary(testutil.Ctx(), memberID)
	if err != nil {
		t.Fatal(err)
	}
	before := beforeSummary.Remaining
	if err := svc.DeletePlatformKey(testutil.Ctx(), created.ID); err != nil {
		t.Fatal(err)
	}
	afterSummary, err := svc.BudgetSummary(testutil.Ctx(), memberID)
	if err != nil {
		t.Fatal(err)
	}
	after := afterSummary.Remaining
	if after <= before {
		t.Fatalf("expected quota release after delete, before=%v after=%v", before, after)
	}
	_ = st
}

func TestRejectApprovalNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	err := svc.RejectApproval(testutil.Ctx(), "missing-approval", contract.IDMemberAdmin, nil)
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestApprovalBudgetCheckNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	_, err := svc.ApprovalBudgetCheck(testutil.Ctx(), "missing-approval")
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestBudgetSummaryIncludesSnapshotUsed(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	testutil.SetPlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1, 1000)
	testutil.SetPlatformKeySnapshotUsed(t, st, "plk-1b", 234.5)
	summary, err := svc.BudgetSummary(ctx, contract.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Used != 1234.5 {
		t.Fatalf("expected used 1234.5 from snapshot, got %v", summary.Used)
	}
}

func TestRevokePlatformKey(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(99)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.RevokePlatformKey(ctx, contract.IDPlatformKey1); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 && key.Status != "revoked" {
			t.Fatalf("expected revoked status, got %s", key.Status)
		}
	}
}
