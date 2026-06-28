package session_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newSessionService(t *testing.T) session.Service {
	t.Helper()
	_, st := testutil.NewMemoryStoreFromConfig(t)
	return session.NewService(st)
}

func TestGetByMemberIDSuccess(t *testing.T) {
	svc := newSessionService(t)
	ctx, err := svc.GetByMemberID(seed.IDMemberAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.ReadOnly {
		t.Fatal("expected admin session to be writable")
	}
	if len(ctx.Permissions) != len(permission.AllPermissions) {
		t.Fatalf("expected %d permissions, got %d", len(permission.AllPermissions), len(ctx.Permissions))
	}
	if ctx.Member.ID != seed.IDMemberAdmin {
		t.Fatalf("expected %s, got %s", seed.IDMemberAdmin, ctx.Member.ID)
	}
}

func TestGetByMemberIDNotFound(t *testing.T) {
	svc := newSessionService(t)
	_, err := svc.GetByMemberID("missing")
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusNotFound {
		t.Fatalf("expected 404, got %v", err)
	}
}

func TestGetByMemberIDReadOnlyMember(t *testing.T) {
	svc := newSessionService(t)
	ctx, err := svc.GetByMemberID("m-pure")
	if err != nil {
		t.Fatal(err)
	}
	if !ctx.ReadOnly {
		t.Fatal("expected m-pure session to be read-only")
	}
}
