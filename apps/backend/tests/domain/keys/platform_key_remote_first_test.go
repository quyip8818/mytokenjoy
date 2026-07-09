package keys_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestTogglePlatformKeyRemoteFailureKeepsStatus(t *testing.T) {
	t.Parallel()
	svc, st, stub := newKeysServiceWithRelay(t)
	ctx := testutil.Ctx()
	tokenID := int64(55)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		SyncStatus:    store.RelaySyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}
	stub.UpdateTokenFn = func(_ context.Context, _ newapi.UpdateTokenRequest) (newapi.Token, error) {
		return newapi.Token{}, errors.New("relay down")
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
		t.Fatal("expected error when relay update fails")
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
	svc, st, stub := newKeysServiceWithRelay(t)
	ctx := testutil.Ctx()
	stub.DeleteTokenFn = func(_ context.Context, _ int64) error {
		return errors.New("relay down")
	}
	mapping := store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: ptrInt64(99),
		SyncStatus:    store.RelaySyncStatusSynced,
	}
	if err := st.Relay().UpsertMapping(ctx, mapping); err != nil {
		t.Fatal(err)
	}
	err := svc.RevokePlatformKey(ctx, contract.IDPlatformKey1)
	if err == nil {
		t.Fatal("expected error when relay delete fails")
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
	svc, st, stub := newKeysServiceWithRelay(t)
	ctx := testutil.Ctx()
	tokenID := int64(77)
	mapping := store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		SyncStatus:    store.RelaySyncStatusSynced,
	}
	if err := st.Relay().UpsertMapping(ctx, mapping); err != nil {
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
	svc, st, stub := newKeysServiceWithRelay(t)
	ctx := testutil.Ctx()
	tokenID := int64(88)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		SyncStatus:    store.RelaySyncStatusSynced,
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

func TestNilRelayClientReturns503(t *testing.T) {
	t.Parallel()
	svc, _ := newKeysService(t)
	memberID := contract.IDMember1
	_, err := svc.TogglePlatformKey(testutil.Ctx(), contract.IDPlatformKey1, false)
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
	_, err = svc.CreatePlatformKey(testutil.Ctx(), types.CreatePlatformKeyInput{
		Name: "x", MemberID: &memberID, Quota: 100,
		ModelWhitelist: []int64{contract.IDModel1},
	})
	testutil.AssertDomainStatus(t, err, domain.StatusServiceUnavailable)
}

func ptrInt64(v int64) *int64 {
	return &v
}
