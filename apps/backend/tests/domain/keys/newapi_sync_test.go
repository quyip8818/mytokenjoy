package keys_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
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
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 881, Key: "sk-test", RemainQuota: 1000}}
	newAPISync, st := newNewAPISync(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	plkTest := uuid.MustParse("00000000-0000-7000-0000-00000000aa01")
	key := types.PlatformKey{
		ID: plkTest, Name: "test-key", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
		Status: "active", Budget: 1000, ModelWhitelist: []uuid.UUID{contract.IDModel1},
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

	if riverfix.ListPendingNewAPISync(t, st, outbox.KindCreateKey, 10) == 0 {
		t.Fatal("expected create_key outbox entry")
	}
}

func TestTrySyncCreateCallsAdminAPI(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 882, Key: "sk-test-key", RemainQuota: 1000}}
	newAPISync, st := newNewAPISync(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	plkSync := uuid.MustParse("00000000-0000-7000-0000-00000000aa02")
	key := types.PlatformKey{
		ID: plkSync, Name: "sync-key", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
		Status: "active", Budget: 1000, ModelWhitelist: []uuid.UUID{contract.IDModel1},
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

	fullKey, err := newAPISync.TrySyncCreate(testutil.Ctx(), plkSync)
	if err != nil {
		t.Fatal(err)
	}
	if fullKey != "sk-test-key" {
		t.Fatalf("expected sk-test-key, got %q", fullKey)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}

	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, plkSync)
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
		if k.ID == plkSync {
			byHash, err := st.Keys().PlatformKeyByHash(ctx, store.HashPlatformKey("sk-test-key"))
			if err != nil || byHash == nil || byHash.ID != plkSync {
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
	plkRollback := uuid.MustParse("00000000-0000-7000-0000-00000000aa03")
	key := types.PlatformKey{
		ID: plkRollback, Name: "rollback", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
		Status: "active", Budget: 500, ModelWhitelist: []uuid.UUID{contract.IDModel1},
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

	newAPISync.RollbackFailedCreate(ctx, plkRollback)

	keys, err = st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		if k.ID == plkRollback {
			t.Fatal("expected plk-rollback removed after rollback")
		}
	}
}
