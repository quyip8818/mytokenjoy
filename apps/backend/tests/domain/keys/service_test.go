package keys_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

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
	svc, _, _ := newKeysServiceWithNewAPI(t)
	memberID := contract.IDMember1
	_, err := svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "too-big", Scope: types.PlatformKeyScopeMember, MemberID: &memberID, Budget: 99_000_000_000,
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
	svc, _, _ := newKeysServiceWithNewAPI(t)
	quota := int64(99_000_000_000)
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
		PlatformKeyID: contract.IDPlatformKey6,
		NewAPIKeyID:   88,
	})

	for _, tc := range []struct {
		name    string
		budget  int64
		wantErr int
	}{
		{name: "rejects over member sub cap", budget: budgetfix.QuotaFromDisplay(7000), wantErr: domain.StatusUnprocessable},
		{name: "allows within member sub cap", budget: budgetfix.QuotaFromDisplay(5500)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := svc.UpdatePlatformKey(ctx, contract.IDPlatformKey6, types.UpdatePlatformKeyInput{Budget: &tc.budget})
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
	newBudget := budgetfix.QuotaFromDisplay(4000)
	if _, err := svc.UpdatePlatformKey(ctx, contract.IDPlatformKey1, types.UpdatePlatformKeyInput{
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

func TestBudgetSummaryIncludesSnapshotConsumed(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	// Zero out existing consumed, then set controlled values.
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1, 1000)
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey3, 234)
	summary, err := svc.BudgetSummary(ctx, contract.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Consumed != 1234 {
		t.Fatalf("expected consumed 1234 from snapshot, got %v", summary.Consumed)
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
	rotated, err := svc.RotateProviderKey(ctx, created.ID, "sk-rotated-new-key")
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
	_, err = svc.RotateProviderKey(ctx, created.ID, "sk-should-fail")
	if err == nil {
		t.Fatal("expected rotate to fail when rotateEnabled=false")
	}
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Status != domain.StatusForbidden {
		t.Fatalf("expected Forbidden, got %v", err)
	}
}
