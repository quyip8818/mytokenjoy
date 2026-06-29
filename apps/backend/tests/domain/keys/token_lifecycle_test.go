package keys_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestSyncCreateEnqueuesOutbox(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: 1000}}
	lifecycle, st := newTokenLifecycle(t, stub)
	memberID := seed.IDMember1
	key := types.PlatformKey{
		ID: "plk-test", Name: "test-key", MemberID: &memberID,
		Status: "active", Quota: 1000, ModelWhitelist: []string{"gpt-4o"},
	}
	keys := st.Keys().PlatformKeys()
	keys = append(keys, key)
	st.Keys().SetPlatformKeys(keys)

	if err := lifecycle.SyncCreatePlatformKey(context.Background(), key, seed.IDDept3); err != nil {
		t.Fatal(err)
	}

	entries, err := st.Relay().ClaimPendingRelayOutbox(10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.Kind == store.OutboxKindCreateToken {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected create_token outbox entry")
	}
}

func TestTrySyncCreateCallsAdminAPI(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, Key: "sk-test-key", RemainQuota: 1000}}
	lifecycle, st := newTokenLifecycle(t, stub)
	memberID := seed.IDMember1
	key := types.PlatformKey{
		ID: "plk-sync", Name: "sync-key", MemberID: &memberID,
		Status: "active", Quota: 1000, ModelWhitelist: []string{"gpt-4o"},
	}
	keys := st.Keys().PlatformKeys()
	keys = append(keys, key)
	st.Keys().SetPlatformKeys(keys)

	if err := lifecycle.SyncCreatePlatformKey(context.Background(), key, seed.IDDept3); err != nil {
		t.Fatal(err)
	}

	fullKey, err := lifecycle.TrySyncCreate(context.Background(), "plk-sync")
	if err != nil {
		t.Fatal(err)
	}
	if fullKey != "sk-test-key" {
		t.Fatalf("expected sk-test-key, got %q", fullKey)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}

	mapping, err := st.Relay().GetMappingByPlatformKeyID("plk-sync")
	if err != nil || mapping == nil {
		t.Fatalf("expected relay mapping, err=%v mapping=%v", err, mapping)
	}
	if mapping.SyncStatus != store.RelaySyncStatusSynced {
		t.Fatalf("expected synced status, got %q", mapping.SyncStatus)
	}

	for _, k := range st.Keys().PlatformKeys() {
		if k.ID == "plk-sync" {
			if k.FullKey == nil || *k.FullKey != "sk-test-key" {
				t.Fatalf("expected full key set, got %+v", k.FullKey)
			}
			return
		}
	}
	t.Fatal("plk-sync not found after sync")
}

func TestRollbackFailedCreate(t *testing.T) {
	stub := &mock.StubAdminClient{}
	lifecycle, st := newTokenLifecycle(t, stub)
	memberID := seed.IDMember1
	key := types.PlatformKey{
		ID: "plk-rollback", Name: "rollback", MemberID: &memberID,
		Status: "active", Quota: 500, ModelWhitelist: []string{"gpt-4o"},
	}
	keys := st.Keys().PlatformKeys()
	keys = append(keys, key)
	st.Keys().SetPlatformKeys(keys)

	lifecycle.RollbackFailedCreate("plk-rollback")

	for _, k := range st.Keys().PlatformKeys() {
		if k.ID == "plk-rollback" {
			t.Fatal("expected plk-rollback removed after rollback")
		}
	}
}
