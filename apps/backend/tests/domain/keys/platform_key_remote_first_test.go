package keys_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestTogglePlatformKeyRemoteFailureKeepsStatus(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(55)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	stub.UpdateTokenFn = func(_ context.Context, _ newapi.UpdateTokenRequest) (newapi.Token, error) {
		return newapi.Token{}, errors.New("newapi down")
	}
	before, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var statusBefore string
	for _, key := range before {
		if key.ID == contract.IDPlatformKey1 {
			statusBefore = key.Status
			break
		}
	}
	_, err = svc.TogglePlatformKey(ctx, contract.IDPlatformKey1, false)
	if err == nil {
		t.Fatal("expected error when newapi update fails")
	}
	after, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range after {
		if key.ID == contract.IDPlatformKey1 && key.Status != statusBefore {
			t.Fatalf("expected status unchanged %q, got %q", statusBefore, key.Status)
		}
	}
}

func TestRevokePlatformKeyRemoteFailureKeepsStatus(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	stub.DeleteTokenFn = func(_ context.Context, _ int64) error {
		return errors.New("newapi down")
	}
	mapping := store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   ptrInt64(99),
		SyncStatus:    store.MappingSyncStatusSynced,
	}
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, mapping); err != nil {
		t.Fatal(err)
	}
	err := svc.RevokePlatformKey(ctx, contract.IDPlatformKey1)
	if err == nil {
		t.Fatal("expected error when newapi delete fails")
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 && key.Status == "revoked" {
			t.Fatal("expected status not revoked after remote failure")
		}
	}
}

func TestRotatePlatformKeySuccess(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(77)
	mapping := store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, mapping); err != nil {
		t.Fatal(err)
	}
	stub.RegenerateTokenFn = func(_ context.Context, id int64) (newapi.Token, error) {
		if id != tokenID {
			t.Fatalf("expected token id %d, got %d", tokenID, id)
		}
		return newapi.Token{ID: id, Key: "sk-rotated-key"}, nil
	}
	rotated, err := svc.RotatePlatformKey(ctx, contract.IDPlatformKey1)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.FullKey == nil || *rotated.FullKey != "sk-rotated-key" {
		t.Fatalf("expected rotated key, got %+v", rotated.FullKey)
	}
	if stub.RegenerateTokenCalls != 1 {
		t.Fatalf("expected one regenerate call, got %d", stub.RegenerateTokenCalls)
	}
}

func TestRotatePlatformKeyRequiresActiveStatus(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(88)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TogglePlatformKey(ctx, contract.IDPlatformKey1, false); err != nil {
		t.Fatal(err)
	}
	stub.RegenerateTokenCalls = 0
	_, err := svc.RotatePlatformKey(ctx, contract.IDPlatformKey1)
	testutil.AssertDomainStatus(t, err, 409)
	if stub.RegenerateTokenCalls != 0 {
		t.Fatalf("expected no regenerate call, got %d", stub.RegenerateTokenCalls)
	}
}

func TestNilNewAPIClientReturns503(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.TogglePlatformKey(testutil.Ctx(), contract.IDPlatformKey1, false)
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
	_, err = svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "x", MemberID: &memberID, Budget: 100,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestDeletePlatformKeyRequiresNewAPI(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	err := svc.DeletePlatformKey(testutil.Ctx(), contract.IDPlatformKey1)
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
}

func TestDeletePlatformKeyRemoteFailureKeepsKey(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	stub.DeleteTokenFn = func(_ context.Context, _ int64) error {
		return errors.New("newapi down")
	}
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   ptrInt64(99),
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	err := svc.DeletePlatformKey(ctx, contract.IDPlatformKey1)
	if err == nil {
		t.Fatal("expected error when newapi delete fails")
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			return
		}
	}
	t.Fatal("expected key to remain after remote failure")
}

func TestDeletePlatformKeyRevokesRemoteToken(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(66)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeletePlatformKey(ctx, contract.IDPlatformKey1); err != nil {
		t.Fatal(err)
	}
	if stub.DeleteTokenCalls != 1 {
		t.Fatalf("expected one delete token call, got %d", stub.DeleteTokenCalls)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			t.Fatal("expected key removed from store")
		}
	}
}

func TestUpdatePlatformKeyRequiresNewAPI(t *testing.T) {
	t.Parallel()
	svc, st := newKeysService(t)
	ctx := testutil.Ctx()
	name := "updated-name"
	_, err := svc.UpdatePlatformKey(ctx, contract.IDPlatformKey1, types.UpdatePlatformKeyInput{
		Name: &name,
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 && key.Name == name {
			t.Fatal("expected name unchanged when newapi disabled")
		}
	}
}

func TestUpdatePlatformKeyRemoteFailureRollsBack(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(55)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	stub.UpdateTokenFn = func(_ context.Context, _ newapi.UpdateTokenRequest) (newapi.Token, error) {
		return newapi.Token{}, errors.New("newapi down")
	}
	before, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var nameBefore string
	for _, key := range before {
		if key.ID == contract.IDPlatformKey1 {
			nameBefore = key.Name
			break
		}
	}
	newName := nameBefore + "-failed"
	_, err = svc.UpdatePlatformKey(ctx, contract.IDPlatformKey1, types.UpdatePlatformKeyInput{
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error when newapi update fails")
	}
	after, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range after {
		if key.ID == contract.IDPlatformKey1 && key.Name != nameBefore {
			t.Fatalf("expected name rollback to %q, got %q", nameBefore, key.Name)
		}
	}
}

func TestUpdatePlatformKeyAppliesSyncImmediately(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithNewAPI(t)
	ctx := testutil.Ctx()
	tokenID := int64(55)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	name := "synced-name"
	_, err := svc.UpdatePlatformKey(ctx, contract.IDPlatformKey1, types.UpdatePlatformKeyInput{
		Name: &name,
	})
	if err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls < 1 {
		t.Fatalf("expected sync update call, got %d", stub.UpdateTokenCalls)
	}
	if pending := riverfix.ListPendingNewAPISync(st, newapisync.OutboxKindUpdateKey, 100); pending != 0 {
		t.Fatalf("expected no pending update_key outbox, got %d", pending)
	}
}
