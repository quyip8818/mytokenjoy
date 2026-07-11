package company_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestSaaSLoginRequiresCompanySlug(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})

	body, _ := json.Marshal(map[string]string{
		"email": "accept@example.com", "password": "securepass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSaaSLoginWithCompanySlug(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	platformCookie := saas.LoginPlatform(t, app.Router)
	created := saas.CreateCompanyHTTP(t, app.Router, platformCookie,
		"login-co", "Login Co", "login@example.com")
	saas.AcceptInviteHTTP(t, app.Router, created.InviteCode, "Login Admin", "securepass123")

	body, _ := json.Marshal(map[string]string{
		"email":       "login@example.com",
		"password":    "securepass123",
		"companySlug": "login-co",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAcceptInviteAssignsInviteRole(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	platformCookie := saas.LoginPlatform(t, app.Router)
	created := saas.CreateCompanyHTTP(t, app.Router, platformCookie,
		"role-co", "Role Co", "role@example.com")
	member, _ := saas.AcceptInviteHTTP(t, app.Router, created.InviteCode, "Role Admin", "securepass123")
	if len(member.Roles) != 1 || member.Roles[0] != permission.RoleSuperAdmin {
		t.Fatalf("expected super admin role, got %v", member.Roles)
	}
}
