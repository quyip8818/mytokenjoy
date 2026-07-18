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

func TestSaaSLoginRequiresCompanyID(t *testing.T) {
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

func TestSaaSLoginWithCompanyID(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	platformCookie := saas.LoginPlatform(t, app.Router)
	created := saas.CreateCompanyHTTP(t, app.Router, platformCookie,
		"Login Co", "login@example.com")
	saas.AcceptInviteHTTP(t, app.Router, created.InviteCode, "Login Admin", "securepass123")

	body, _ := json.Marshal(map[string]any{
		"email":     "login@example.com",
		"password":  "securepass123",
		"companyId": created.Company.ID,
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
		"Role Co", "role@example.com")
	_, memberCookie := saas.AcceptInviteHTTP(t, app.Router, created.InviteCode, "Role Admin", "securepass123")

	// Verify role via session endpoint.
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", memberCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var session struct {
		Member struct {
			Roles []string `json:"roles"`
		} `json:"member"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if len(session.Member.Roles) != 1 || session.Member.Roles[0] != permission.RoleSuperAdmin {
		t.Fatalf("expected super admin role, got %v", session.Member.Roles)
	}
}
