package newapisync_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestTrySyncCreateEnsuresGroupBeforeCreateToken(t *testing.T) {
	testutil.SkipUnlessLocal(t)
	t.Parallel()
	var ensuredGroup string
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 991, Key: "sk-test-key", RemainQuota: 1000},
		EnsureGroupFn: func(_ context.Context, group, _ string) error {
			ensuredGroup = group
			return nil
		},
	}
	sync, st := newSyncWithStub(t, stub)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	plkEnsureGroup := uuid.MustParse("00000000-0000-7000-0000-00000000ff01")
	key := types.PlatformKey{
		ID: plkEnsureGroup, Name: "ensure-group", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
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
	if err := sync.SyncCreatePlatformKey(ctx, key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}
	if _, err := sync.TrySyncCreate(ctx, plkEnsureGroup); err != nil {
		t.Fatal(err)
	}
	wantGroup := contract.DefaultCompanyID.String() // company-level group, not per-department
	if stub.EnsureGroupCalls != 1 {
		t.Fatalf("expected one EnsureGroup call, got %d", stub.EnsureGroupCalls)
	}
	if ensuredGroup != wantGroup {
		t.Fatalf("expected group %q, got %q", wantGroup, ensuredGroup)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}
}
