package authz_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newAuthzService(t *testing.T) authz.Service {
	t.Helper()
	cfg := testutil.TestConfig()
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	return authz.NewService(cfg, st)
}

func TestGetSessionContextSuccess(t *testing.T) {
	t.Parallel()
	svc := newAuthzService(t)
	ctx, err := svc.GetSessionContext(testutil.Ctx(), contract.DefaultCompanyID, contract.IDMemberAdmin)
	if err != nil {
		t.Fatalf("GetSessionContext: %v", err)
	}
	if ctx.Member.ID != contract.IDMemberAdmin {
		t.Fatalf("expected member %s, got %s", contract.IDMemberAdmin, ctx.Member.ID)
	}
	if len(ctx.Permissions) == 0 {
		t.Fatal("expected permissions")
	}
	if ctx.AuthzRevision < 0 {
		t.Fatal("expected authz revision")
	}
}

func TestGetSessionContextNotFound(t *testing.T) {
	t.Parallel()
	svc := newAuthzService(t)
	_, err := svc.GetSessionContext(testutil.Ctx(), contract.DefaultCompanyID, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSessionContextReadOnlyMember(t *testing.T) {
	t.Parallel()
	svc := newAuthzService(t)
	ctx, err := svc.GetSessionContext(testutil.Ctx(), contract.DefaultCompanyID, "m-pure")
	if err != nil {
		t.Fatalf("GetSessionContext: %v", err)
	}
	if !ctx.ReadOnly {
		t.Fatal("expected read-only session")
	}
}
