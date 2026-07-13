//go:build testhook

package newapisync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestResolvePlatformKeyBearerReturnsSecretWithoutRotate(t *testing.T) {
	t.Parallel()

	tokenID := int64(77)
	const bearer = "sk-test-key"
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: tokenID, Key: bearer, RemainQuota: 1000},
		GetTokenKeyFn: func(_ context.Context, id int64) (string, error) {
			if id != tokenID {
				t.Fatalf("unexpected token id %d", id)
			}
			return bearer, nil
		},
	}

	sync, st := newSyncWithStub(t, stub)
	ctx := testutil.Ctx()

	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		DepartmentID:  contract.IDDept3,
		NewAPIGroup:   "dept-dept-3",
		NewAPIKeyID:   &tokenID,
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}

	got, err := sync.ResolvePlatformKeyBearer(ctx, contract.IDPlatformKey1)
	if err != nil {
		t.Fatal(err)
	}
	if got != bearer {
		t.Fatalf("expected bearer %q, got %q", bearer, got)
	}
	if stub.UpdateTokenCalls != 0 {
		t.Fatalf("expected no UpdateToken calls, got %d", stub.UpdateTokenCalls)
	}
	if stub.GetTokenKeyCalls != 1 {
		t.Fatalf("expected one GetTokenKey call, got %d", stub.GetTokenKeyCalls)
	}
	if stub.RegenerateTokenCalls != 0 {
		t.Fatalf("expected no regenerate call, got %d", stub.RegenerateTokenCalls)
	}
}
