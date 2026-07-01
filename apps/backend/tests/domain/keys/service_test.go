package keys_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApprovalQuotaCheckInsufficient(t *testing.T) {
	svc, _ := newKeysService(t)
	check, err := svc.ApprovalQuotaCheck(testutil.Ctx(), seed.IDApproval1)
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

func TestApprovalQuotaCheckSufficient(t *testing.T) {
	svc, st := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "quota", Reason: "test", RequestedQuota: 1000,
		RequestedModels: []string{"gpt-4o"}, MemberID: seed.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	check, err := svc.ApprovalQuotaCheck(testutil.Ctx(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !check.Sufficient {
		t.Fatalf("expected sufficient, reserved=%v requested=%v", check.ReservedPool, check.Requested)
	}
	_ = st
}

func TestApproveKeyTypeCreatesPlatformKey(t *testing.T) {
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	keysBefore, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before := len(keysBefore)
	if err := svc.ApproveApproval(testutil.Ctx(), seed.IDApproval1, seed.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	approval := findApproval(st, seed.IDApproval1)
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

func TestApproveQuotaTypeAddsPersonalQuota(t *testing.T) {
	svc, st := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "quota", Reason: "need more", RequestedQuota: 1000,
		RequestedModels: []string{"gpt-4o"}, MemberID: seed.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	poolBefore, err := st.Budget().MemberQuotaPools(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	before := poolBefore[seed.IDMember1].PersonalQuota
	if err := svc.ApproveApproval(testutil.Ctx(), created.ID, seed.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	poolAfter, err := st.Budget().MemberQuotaPools(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	after := poolAfter[seed.IDMember1].PersonalQuota
	if after != before+1000 {
		t.Fatalf("expected personal quota +1000, before=%v after=%v", before, after)
	}
}

func TestApproveInsufficientReserved(t *testing.T) {
	svc, _ := newKeysService(t)
	created, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "quota", Reason: "too much", RequestedQuota: 9999,
		RequestedModels: []string{"gpt-4o"}, MemberID: seed.IDMember1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = svc.ApproveApproval(testutil.Ctx(), created.ID, seed.IDMemberAdmin)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestRejectApproval(t *testing.T) {
	svc, st := newKeysService(t)
	reason := "not needed"
	if err := svc.RejectApproval(testutil.Ctx(), seed.IDApproval2, seed.IDMemberAdmin, &reason); err != nil {
		t.Fatal(err)
	}
	approval := findApproval(st, seed.IDApproval2)
	if approval == nil || approval.Status != "rejected" {
		t.Fatalf("expected apv-2 rejected, got %+v", approval)
	}
}

func TestCreatePlatformKeySuccess(t *testing.T) {
	svc, _ := newKeysService(t)
	memberID := seed.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "test-key", MemberID: &memberID, Quota: 1000,
		ModelWhitelist: []string{"gpt-4o"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.FullKey == nil || !strings.Contains(*created.FullKey, common.DemoKeyPrefix) {
		t.Fatalf("expected demo full key with prefix %q, got %+v", common.DemoKeyPrefix, created.FullKey)
	}
}

func TestCreatePlatformKeyQuotaExceeded(t *testing.T) {
	svc, _ := newKeysService(t)
	memberID := seed.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "too-big", MemberID: &memberID, Quota: 99999,
		ModelWhitelist: []string{"gpt-4o"},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreatePlatformKeyInvalidWhitelist(t *testing.T) {
	svc, _ := newKeysService(t)
	memberID := seed.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "bad-models", MemberID: &memberID, Quota: 1000,
		ModelWhitelist: []string{"nonexistent-model"},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateApprovalInvalidModels(t *testing.T) {
	svc, _ := newKeysService(t)
	_, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "quota", Reason: "bad models", RequestedQuota: 1000,
		RequestedModels: []string{"nonexistent-model"}, MemberID: seed.IDMember1,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateGroupKeyQuotaExceeded(t *testing.T) {
	svc, _ := newKeysService(t)
	groupID := seed.IDBudgetGroup1
	memberID := seed.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "group-over", BudgetGroupID: &groupID, MemberID: &memberID, Quota: 99999,
		ModelWhitelist: []string{"gpt-4o"},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdatePlatformKeyQuota(t *testing.T) {
	svc, _ := newKeysService(t)
	quota := 99999.0
	_, err := svc.UpdatePlatformKey(testutil.Ctx(), seed.IDPlatformKey1, types.UpdatePlatformKeyInput{
		Quota: &quota,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestDeletePlatformKeyReleasesQuota(t *testing.T) {
	svc, st := newKeysService(t)
	memberID := seed.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "release-me", MemberID: &memberID, Quota: 500,
		ModelWhitelist: []string{"gpt-4o"},
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeSummary, err := svc.QuotaSummary(testutil.Ctx(), memberID)
	if err != nil {
		t.Fatal(err)
	}
	before := beforeSummary.Remaining
	if err := svc.DeletePlatformKey(testutil.Ctx(), created.ID); err != nil {
		t.Fatal(err)
	}
	afterSummary, err := svc.QuotaSummary(testutil.Ctx(), memberID)
	if err != nil {
		t.Fatal(err)
	}
	after := afterSummary.Remaining
	if after <= before {
		t.Fatalf("expected quota release after delete, before=%v after=%v", before, after)
	}
	_ = st
}

func TestRevokePlatformKey(t *testing.T) {
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	if err := svc.RevokePlatformKey(testutil.Ctx(), seed.IDPlatformKey1); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == seed.IDPlatformKey1 && key.Status != "revoked" {
			t.Fatalf("expected revoked status, got %s", key.Status)
		}
	}
}
