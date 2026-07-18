package keys_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
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
		RequestedModels: []uuid.UUID{contract.IDModel1}, MemberID: contract.IDMember1,
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

func TestApproveKeySyncFailureRevertsApproval(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	stub.CreateTokenFn = func(context.Context, newapi.CreateTokenRequest) (newapi.Token, error) {
		return newapi.Token{}, errors.New("newapi create failed")
	}
	keysBefore, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before := len(keysBefore)

	err = svc.ApproveApproval(ctx, contract.IDApproval1, contract.IDMemberAdmin)
	if err == nil {
		t.Fatal("expected approve to fail when newapi create fails")
	}

	approval := findApproval(st, contract.IDApproval1)
	if approval == nil || approval.Status != "pending" {
		t.Fatalf("expected apv-1 reverted to pending, got %+v", approval)
	}
	after, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != before {
		t.Fatalf("expected no persisted platform key after sync failure, before=%d after=%d", before, len(after))
	}
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
		RequestedModels: []uuid.UUID{contract.IDModel1}, MemberID: contract.IDMember1,
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
		RequestedModels: []uuid.UUID{contract.IDModel1}, MemberID: contract.IDMember1,
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
		Name: "test-key", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
}

func TestCreatePlatformKeySuccess(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	memberID := contract.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "test-key", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.FullKey == nil || *created.FullKey != "sk-test-key" {
		t.Fatalf("expected platform key full key, got %+v", created.FullKey)
	}
	remain, _ := budgetfix.CombinedKeyRemain(t, st, created.ID)
	if remain == nil {
		t.Fatal("expected gateway soft remain after key create")
	}
}

func TestCreatePlatformKeyQuotaExceeded(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "too-big", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 20_000_000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreatePlatformKeyInvalidWhitelist(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "bad-models", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 1000,
		ModelWhitelist: []uuid.UUID{uuid.MustParse("00000000-0000-7000-0000-0000000f4240")},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateApprovalInvalidModels(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	_, err := svc.CreateApproval(testutil.Ctx(), types.CreateApprovalInput{
		Type: "budget", Reason: "bad models", RequestedBudget: 1000,
		RequestedModels: []uuid.UUID{uuid.MustParse("00000000-0000-7000-0000-0000000f4240")}, MemberID: contract.IDMember1,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestCreateProjectKeyQuotaExceeded(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	groupID := contract.IDProject1
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "group-over", Scope: types.PlatformKeyScopeProject, ProjectID: &groupID, MemberID: &memberID, Budget: 20_000_000,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
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

func TestUpdatePlatformKeyProjectMemberBudget(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	budgetfix.SetProjectSnapshotConsumed(t, st, contract.IDProject1, 0)
	newapisynctf.UpsertMapping(t, st, newapisynctf.MappingOpts{
		PlatformKeyID: uuid.MustParse("00000000-0000-7000-0000-00000000b901"),
		NewAPIKeyID:   88,
	})

	for _, tc := range []struct {
		name    string
		budget  float64
		wantErr int
	}{
		{name: "rejects over member sub cap", budget: budgetfix.DisplayPoints(7000), wantErr: domain.StatusUnprocessable},
		{name: "allows within member sub cap", budget: budgetfix.DisplayPoints(5500)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := svc.UpdatePlatformKey(ctx, "00000000-0000-7000-0000-00000000b901", types.UpdatePlatformKeyInput{Budget: &tc.budget})
			if tc.wantErr != 0 {
				testutil.AssertDomainStatus(t, err, tc.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if updated.Budget != tc.budget {
				t.Fatalf("expected budget %v, got %v", tc.budget, updated.Budget)
			}
		})
	}
}

func TestUpdatePlatformKeyRefreshesGatewaySoft(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	versionBefore := budgetfix.CombinedKeyRemainVersion(t, st, contract.IDPlatformKey1)
	newBudget := budgetfix.DisplayPoints(4000)
	if _, err := svc.UpdatePlatformKey(ctx, contract.IDPlatformKey1.String(), types.UpdatePlatformKeyInput{
		Budget: &newBudget,
	}); err != nil {
		t.Fatal(err)
	}
	versionAfter := budgetfix.CombinedKeyRemainVersion(t, st, contract.IDPlatformKey1)
	if versionAfter <= versionBefore {
		t.Fatalf("expected gateway soft version increase, before=%d after=%d", versionBefore, versionAfter)
	}
}

func TestDeletePlatformKeyReleasesQuota(t *testing.T) {
	t.Parallel()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	memberID := contract.IDMember1
	created, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "release-me", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 500,
		ModelWhitelist: []uuid.UUID{contract.IDModel1},
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeSummary, err := svc.BudgetSummary(testutil.Ctx(), memberID)
	if err != nil {
		t.Fatal(err)
	}
	before := beforeSummary.Remaining
	if err := svc.DeletePlatformKey(testutil.Ctx(), created.ID.String()); err != nil {
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
	err := svc.RejectApproval(testutil.Ctx(), "missing-approval", contract.IDMemberAdmin.String(), nil)
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestApprovalBudgetCheckNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	_, err := svc.ApprovalBudgetCheck(testutil.Ctx(), "00000000-0000-7000-8000-ffffffffffff")
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestBudgetSummaryIncludesSnapshotConsumed(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1, 1000)
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, uuid.MustParse("00000000-0000-7000-0000-00000000f01b"), 234.5)
	summary, err := svc.BudgetSummary(ctx, contract.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Consumed != 1234.5 {
		t.Fatalf("expected consumed 1234.5 from snapshot, got %v", summary.Consumed)
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
	if err := svc.RevokePlatformKey(ctx, contract.IDPlatformKey1.String()); err != nil {
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

func TestRotateProviderKeyRespectsRotateEnabled(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	created, err := svc.CreateProviderKey(ctx, types.CreateProviderKeyInput{
		Provider: "openai", Name: "rot-test", Key: "sk-rot-enabled-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created.RotateEnabled {
		t.Fatal("expected newly created provider keys to allow rotation")
	}
	rotated, err := svc.RotateProviderKey(ctx, created.ID.String(), "sk-rotated-new-key")
	if err != nil {
		t.Fatal(err)
	}
	if rotated.KeyPrefix == created.KeyPrefix {
		t.Fatal("expected key prefix to change after rotate")
	}

	keys, err := st.Keys().ProviderKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == created.ID {
			keys[i].RotateEnabled = false
			if err := st.Keys().SetProviderKeys(ctx, keys); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	_, err = svc.RotateProviderKey(ctx, created.ID.String(), "sk-should-fail")
	if err == nil {
		t.Fatal("expected rotate to fail when rotateEnabled=false")
	}
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Status != domain.StatusForbidden {
		t.Fatalf("expected Forbidden, got %v", err)
	}
}
