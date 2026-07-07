package credentials_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAuthenticateMember_Success(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@test.com", "admin123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	member, err := svc.AuthenticateMember(ctx, seed.DefaultCompanyID, "zhangsan@example.com", seed.DemoPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.ID != seed.IDMember1 {
		t.Errorf("member ID = %q, want %q", member.ID, seed.IDMember1)
	}
}

func TestAuthenticateMember_WrongPassword(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	_, err := svc.AuthenticateMember(ctx, seed.DefaultCompanyID, "zhangsan@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestAuthenticateMember_NonexistentEmail(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	_, err := svc.AuthenticateMember(ctx, seed.DefaultCompanyID, "nobody@example.com", seed.DemoPassword)
	if err == nil {
		t.Fatal("expected error for non-existent email")
	}
}

func TestBootstrapPlatformIfNeeded_CreatesOperator(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatalf("bootstrap error: %v", err)
	}

	// Verify operator was created
	opID, err := svc.AuthenticatePlatform(ctx, "admin@platform.com", "secret123")
	if err != nil {
		t.Fatalf("auth platform error: %v", err)
	}
	if opID == "" {
		t.Error("expected non-empty operator ID")
	}
}

func TestBootstrapPlatformIfNeeded_SkipsWhenOperatorsExist(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	// First bootstrap creates the operator
	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	// Second bootstrap should be a no-op (no duplicate)
	cfg.PlatformBootstrapEmail = "admin2@platform.com"
	cfg.PlatformBootstrapPassword = "other"
	svc2 := credentials.NewService(cfg, st)
	if err := svc2.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	// admin2 should not exist
	_, err := svc2.AuthenticatePlatform(ctx, "admin2@platform.com", "other")
	if err == nil {
		t.Error("expected admin2 to not exist")
	}
}

func TestAuthenticatePlatform_WrongPassword(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithPlatformBootstrap("admin@platform.com", "secret123"),
	)
	svc := credentials.NewService(cfg, st)
	ctx := testutil.Ctx()

	if err := svc.BootstrapPlatformIfNeeded(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := svc.AuthenticatePlatform(ctx, "admin@platform.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
