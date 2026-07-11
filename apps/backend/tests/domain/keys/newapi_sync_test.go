package keys_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestSyncCreateEnqueuesOutbox(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: 1000}}
	newAPISync, st := newNewAPISync(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-test", Name: "test-key", MemberID: &memberID,
		Status: "active", Budget: 1000, ModelWhitelist: []int64{contract.IDModel1},
		CreatedAt: "2026-06-19",
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	if err := newAPISync.SyncCreatePlatformKey(testutil.Ctx(), key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}

	if riverfix.ListPendingNewAPISync(st, newapisync.OutboxKindCreateKey, 10) == 0 {
		t.Fatal("expected create_key outbox entry")
	}
}

func TestTrySyncCreateCallsAdminAPI(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, Key: "sk-test-key", RemainQuota: 1000}}
	newAPISync, st := newNewAPISync(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-sync", Name: "sync-key", MemberID: &memberID,
		Status: "active", Budget: 1000, ModelWhitelist: []int64{contract.IDModel1},
		CreatedAt: "2026-06-19",
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	if err := newAPISync.SyncCreatePlatformKey(testutil.Ctx(), key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}

	fullKey, err := newAPISync.TrySyncCreate(testutil.Ctx(), "plk-sync")
	if err != nil {
		t.Fatal(err)
	}
	if fullKey != "sk-test-key" {
		t.Fatalf("expected sk-test-key, got %q", fullKey)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}

	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, "plk-sync")
	if err != nil || mapping == nil {
		t.Fatalf("expected platform key mapping, err=%v mapping=%v", err, mapping)
	}
	if mapping.SyncStatus != store.MappingSyncStatusSynced {
		t.Fatalf("expected synced status, got %q", mapping.SyncStatus)
	}

	keys, err = st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		if k.ID == "plk-sync" {
			byHash, err := st.Keys().PlatformKeyByHash(ctx, store.HashPlatformKey("sk-test-key"))
			if err != nil || byHash == nil || byHash.ID != "plk-sync" {
				t.Fatalf("expected key hash lookup, err=%v key=%v", err, byHash)
			}
			return
		}
	}
	t.Fatal("plk-sync not found after sync")
}

func TestRollbackFailedCreate(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{}
	newAPISync, st := newNewAPISync(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-rollback", Name: "rollback", MemberID: &memberID,
		Status: "active", Budget: 500, ModelWhitelist: []int64{contract.IDModel1},
		CreatedAt: "2026-06-19",
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	newAPISync.RollbackFailedCreate(ctx, "plk-rollback")

	keys, err = st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		if k.ID == "plk-rollback" {
			t.Fatal("expected plk-rollback removed after rollback")
		}
	}
}
