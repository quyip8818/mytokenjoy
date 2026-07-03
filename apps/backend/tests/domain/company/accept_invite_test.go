package company_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/memory"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestAcceptInviteCreatesSessionReadyMember(t *testing.T) {
	mock := saas.StartNewAPIMock(t)
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	router := app.Router
	platformCookie := saas.LoginPlatform(t, router)
	created := saas.CreateCompanyHTTP(t, router, platformCookie,
		"accept-co", "Accept Co", "accept@example.com")
	_, memberCookie := saas.AcceptInviteHTTP(t, router, created.InviteToken,
		"Accept Admin", "securepass123")

	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", memberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var session struct {
		CompanyID int64 `json:"companyId"`
		Member    struct {
			Email string `json:"email"`
		} `json:"member"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if session.CompanyID != created.Company.ID {
		t.Fatalf("expected company %d, got %d", created.Company.ID, session.CompanyID)
	}
	if session.Member.Email != "accept@example.com" {
		t.Fatalf("expected invite email, got %s", session.Member.Email)
	}
}

func TestAcceptInviteRejectsShortPassword(t *testing.T) {
	cfg := testutil.TestConfig()
	st := memory.New(seed.Load(cfg))
	svc := company.NewService(cfg, st, nil)
	ctx := context.Background()

	now := time.Now().UTC()
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: "invite-test-2", CompanyID: seed.DefaultCompanyID,
		Email: "short@newco.example", Role: store.InviteRoleSuperAdmin,
		Token: "short-token", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		Token: "short-token", Name: "Admin", Password: "short",
	})
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestAcceptInviteRejectsExpiredToken(t *testing.T) {
	cfg := testutil.TestConfig()
	st := memory.New(seed.Load(cfg))
	svc := company.NewService(cfg, st, nil)
	ctx := context.Background()

	now := time.Now().UTC()
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: "invite-expired", CompanyID: seed.DefaultCompanyID,
		Email: "expired@example.com", Role: store.InviteRoleSuperAdmin,
		Token: "expired-token", ExpiresAt: now.Add(-time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		Token: "expired-token", Name: "Admin", Password: "securepass123",
	})
	if err == nil {
		t.Fatal("expected error for expired invite")
	}
}

func TestAcceptInviteRejectsAlreadyAccepted(t *testing.T) {
	cfg := testutil.TestConfig()
	st := memory.New(seed.Load(cfg))
	svc := company.NewService(cfg, st, nil)
	ctx := context.Background()

	now := time.Now().UTC()
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: "invite-used", CompanyID: seed.DefaultCompanyID,
		Email: "used@example.com", Role: store.InviteRoleSuperAdmin,
		Token: "used-token", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		Token: "used-token", Name: "Admin", Password: "securepass123",
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		Token: "used-token", Name: "Other", Password: "securepass456",
	})
	if err == nil {
		t.Fatal("expected error for already accepted invite")
	}
}

func TestAcceptInviteRejectsInvalidToken(t *testing.T) {
	cfg := testutil.TestConfig()
	st := memory.New(seed.Load(cfg))
	svc := company.NewService(cfg, st, nil)

	_, err := svc.AcceptInvite(context.Background(), company.AcceptInviteRequest{
		Token: "does-not-exist", Name: "Admin", Password: "securepass123",
	})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
