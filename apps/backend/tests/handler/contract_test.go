package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

type getContractCase struct {
	name string
	path string
}

func getContractCases() []getContractCase {
	return []getContractCase{
		{name: "healthz", path: "/healthz"},
		{name: "session", path: "/api/session"},
		{name: "org data source status", path: "/api/org/data-source/status"},
		{name: "org data source search", path: "/api/org/data-source/search?keyword="},
		{name: "org sync config", path: "/api/org/sync/config"},
		{name: "org sync logs", path: "/api/org/sync/logs?page=1&pageSize=10"},
		{name: "org departments tree", path: "/api/org/departments/tree"},
		{name: "org members", path: "/api/org/members?page=1&pageSize=20"},
		{name: "org roles", path: "/api/org/roles"},
		{name: "org role members", path: "/api/org/roles/role-1/members"},
		{name: "org permissions", path: "/api/org/permissions"},
		{name: "budget tree", path: "/api/budget/tree"},
		{name: "budget member quotas", path: "/api/budget/departments/dept-3/member-quotas"},
		{name: "budget groups", path: "/api/budget/groups"},
		{name: "budget overrun policy", path: "/api/budget/overrun-policy"},
		{name: "budget alerts", path: "/api/budget/alerts"},
		{name: "keys provider", path: "/api/keys/provider"},
		{name: "keys platform", path: "/api/keys/platform"},
		{name: "keys platform quota summary", path: "/api/keys/platform/quota-summary?memberId=m-1"},
		{name: "keys approvals", path: "/api/keys/approvals"},
		{name: "keys approval quota check", path: "/api/keys/approvals/apv-1/quota-check"},
		{name: "models list", path: "/api/models"},
		{name: "models routing", path: "/api/models/routing"},
		{name: "models routing resolve", path: "/api/models/routing/resolve?deptId=dept-3"},
		{name: "dashboard cost summary", path: "/api/dashboard/cost/summary"},
		{name: "dashboard cost departments", path: "/api/dashboard/cost/departments"},
		{name: "dashboard cost department members", path: "/api/dashboard/cost/departments/dept-3/members"},
		{name: "dashboard cost daily", path: "/api/dashboard/cost/daily"},
		{name: "dashboard cost top", path: "/api/dashboard/cost/top?limit=5"},
		{name: "dashboard usage models", path: "/api/dashboard/usage/models"},
		{name: "dashboard usage teams", path: "/api/dashboard/usage/teams"},
		{name: "dashboard usage series", path: "/api/dashboard/usage/series?granularity=day&start=2026-06-01&end=2026-06-07"},
		{name: "audit settings", path: "/api/audit/settings"},
		{name: "audit operations", path: "/api/audit/operations?page=1&pageSize=20"},
		{name: "audit calls", path: "/api/audit/calls?page=1&pageSize=20"},
	}
}

func TestGetContractEndpoints(t *testing.T) {
	router := newTestRouter(t)
	for _, tc := range getContractCases() {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.path != "/healthz" {
				req.Header.Set("Cookie", adminSessionCookie(t))
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}
			var payload any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("expected JSON body: %v", err)
			}
		})
	}
}

func TestSessionUnauthorizedWithoutCookie(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func saasContractRouter(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	mock := testutil.StartNewAPIMock(t)
	app := newTestApp(t, func(cfg *config.Config) {
		testutil.ApplySaaSConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	router := app.Router
	platformCookie := testutil.LoginPlatform(t, router)
	provisioned := testutil.ProvisionCompanyHTTP(t, router, platformCookie,
		"contract-co", "Contract Co", "contract@example.com", "Contract Admin", "securepass123")
	return router, platformCookie, provisioned.MemberCookie
}

func TestSaaSContractEndpoints(t *testing.T) {
	router, platformCookie, memberCookie := saasContractRouter(t)

	cases := []struct {
		name       string
		method     string
		path       string
		cookie     string
		wantStatus int
	}{
		{name: "platform companies", method: http.MethodGet, path: "/api/platform/companies", cookie: platformCookie, wantStatus: http.StatusOK},
		{name: "billing wallet", method: http.MethodGet, path: "/api/billing/wallet", cookie: memberCookie, wantStatus: http.StatusOK},
		{name: "platform unauthorized", method: http.MethodGet, path: "/api/platform/companies", cookie: "", wantStatus: http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.cookie != "" {
				req.Header.Set("Cookie", tc.cookie)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d body=%s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}

	body := []byte(`{"token":"invalid","name":"X","password":"securepass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("accept-invite invalid token: expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}
