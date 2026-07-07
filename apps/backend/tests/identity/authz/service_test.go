package authz_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newAuthzService(t *testing.T) authz.Service {
	t.Helper()
	cfg := testutil.TestConfig()
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	return authz.NewService(cfg, st)
}

func TestGetSessionContextSuccess(t *testing.T) {
	svc := newAuthzService(t)
	ctx, err := svc.GetSessionContext(testutil.Ctx(), seed.DefaultCompanyID, seed.IDMemberAdmin)
	if err != nil {
		t.Fatalf("GetSessionContext: %v", err)
	}
	if ctx.Member.ID != seed.IDMemberAdmin {
		t.Fatalf("expected member %s, got %s", seed.IDMemberAdmin, ctx.Member.ID)
	}
	if len(ctx.Permissions) == 0 {
		t.Fatal("expected permissions")
	}
	if ctx.AuthzRevision < 0 {
		t.Fatal("expected authz revision")
	}
}

func TestGetSessionContextNotFound(t *testing.T) {
	svc := newAuthzService(t)
	_, err := svc.GetSessionContext(testutil.Ctx(), seed.DefaultCompanyID, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSessionContextReadOnlyMember(t *testing.T) {
	svc := newAuthzService(t)
	ctx, err := svc.GetSessionContext(testutil.Ctx(), seed.DefaultCompanyID, "m-pure")
	if err != nil {
		t.Fatalf("GetSessionContext: %v", err)
	}
	if !ctx.ReadOnly {
		t.Fatal("expected read-only session")
	}
}
