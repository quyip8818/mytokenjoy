package credentials_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAuthenticateMember_Success(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@test.com", "admin123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	member, err := svc.AuthenticateMember(ctx, contract.DefaultCompanyID, "zhangsan@example.com", contract.DemoPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.ID != contract.IDMember1 {
		t.Errorf("member ID = %q, want %q", member.ID, contract.IDMember1)
	}
}

func TestAuthenticateMember_WrongPassword(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	_, err := svc.AuthenticateMember(ctx, contract.DefaultCompanyID, "zhangsan@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestAuthenticateMember_NonexistentEmail(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	_, err := svc.AuthenticateMember(ctx, contract.DefaultCompanyID, "nobody@example.com", contract.DemoPassword)
	if err == nil {
		t.Fatal("expected error for non-existent email")
	}
}

func TestBootstrapPlatformIfNeeded_CreatesAdmin(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatalf("bootstrap error: %v", err)
	}

	// Verify platform admin was created as a member of the super company
	member, err := svc.AuthenticateMember(ctx, cfg.TokenJoyCompanyID, "admin@platform.com", "secret123")
	if err != nil {
		t.Fatalf("auth member error: %v", err)
	}
	if member.ID.String() == "" {
		t.Error("expected non-empty member ID")
	}
}

func TestBootstrapPlatformIfNeeded_SkipsWhenAlreadyExists(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	// First bootstrap creates the admin
	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	// Second bootstrap should be a no-op (idempotent)
	cfg.PlatformBootstrapEmail = "admin@platform.com"
	cfg.PlatformBootstrapPassword = "other-password"
	svc2 := credentials.NewService(cfg, st)
	if err := svc2.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	// Original password should still work (not overwritten)
	_, err := svc.AuthenticateMember(ctx, cfg.TokenJoyCompanyID, "admin@platform.com", "secret123")
	if err != nil {
		t.Fatalf("expected original password to still work: %v", err)
	}
}

func TestBootstrapPlatformIfNeeded_WrongPassword(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := svc.AuthenticateMember(ctx, cfg.TokenJoyCompanyID, "admin@platform.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
