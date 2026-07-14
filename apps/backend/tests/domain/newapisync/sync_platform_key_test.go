//go:build testhook

package newapisync_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestSyncPlatformKeyCreateDoesNotEnqueueOutbox(t *testing.T) {
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 56, Key: "sk-no-outbox", RemainQuota: 1000},
	}
	sync, st := newSyncWithStub(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-no-outbox", Name: "no-outbox", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
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
	before := riverfix.ListPendingNewAPISync(t, st, outbox.KindCreateKey, 100)
	if _, err := sync.SyncPlatformKeyCreate(ctx, key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}
	after := riverfix.ListPendingNewAPISync(t, st, outbox.KindCreateKey, 100)
	if after > before {
		t.Fatalf("expected no new create_key outbox entry, before=%d after=%d", before, after)
	}
}

func TestSyncPlatformKeyCreatePersistsHashAndMapping(t *testing.T) {
	t.Parallel()
	tokenID := int64(55)
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: tokenID, Key: "sk-shared-create", RemainQuota: 1000},
	}
	sync, st := newSyncWithStub(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-shared", Name: "shared-create", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
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

	fullKey, err := sync.SyncPlatformKeyCreate(ctx, key, contract.IDDept3)
	if err != nil {
		t.Fatal(err)
	}
	if fullKey != "sk-shared-create" {
		t.Fatalf("expected sk-shared-create, got %q", fullKey)
	}
	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, key.ID)
	if err != nil {
		t.Fatal(err)
	}
	if mapping == nil || mapping.SyncStatus != store.MappingSyncStatusSynced {
		t.Fatalf("expected synced mapping, got %+v", mapping)
	}
	hash, ok, err := st.Keys().PlatformKeyHashByID(ctx, key.ID)
	if err != nil || !ok || hash != store.HashPlatformKey(fullKey) {
		t.Fatalf("expected persisted hash, ok=%v err=%v", ok, err)
	}
}
