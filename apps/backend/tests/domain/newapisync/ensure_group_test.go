package newapisync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestTrySyncCreateEnsuresGroupBeforeCreateToken(t *testing.T) {
	t.Parallel()
	var ensuredGroup, ensuredDisplay string
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 991, Key: "sk-test-key", RemainQuota: 1000},
		EnsureGroupFn: func(_ context.Context, group, displayName string) error {
			ensuredGroup = group
			ensuredDisplay = displayName
			return nil
		},
	}
	sync, st := newSyncWithStub(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-ensure-group", Name: "ensure-group", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
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
	if err := sync.SyncCreatePlatformKey(ctx, key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}
	if _, err := sync.TrySyncCreate(ctx, "plk-ensure-group"); err != nil {
		t.Fatal(err)
	}
	wantGroup := newapiunits.NewAPIGroupForDepartment(contract.IDDept3)
	if stub.EnsureGroupCalls != 1 {
		t.Fatalf("expected one EnsureGroup call, got %d", stub.EnsureGroupCalls)
	}
	if ensuredGroup != wantGroup {
		t.Fatalf("expected group %q, got %q", wantGroup, ensuredGroup)
	}
	if ensuredDisplay != "后端组" {
		t.Fatalf("expected display %q, got %q", "后端组", ensuredDisplay)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}
}
