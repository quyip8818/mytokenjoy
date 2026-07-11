package authz_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestBareMemberIDCookieRejected(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", "tokenjoy_session_member=m-admin", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBareBearerMemberIDRejected(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", "", "", map[string]string{
		"Authorization": "Bearer m-admin",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTamperedJWTRejected(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	token := testutil.IssueSessionJWT(t, contract.DefaultCompanyID, contract.IDMemberAdmin)
	tampered := token[:len(token)-1] + "x"
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/session", "tokenjoy_session_member="+tampered, "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestJWTCompanyMismatchRejected(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testutil.SessionCookieForCompany(t, 999, contract.IDMemberAdmin)
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/org/departments/tree", cookie, "", nil)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 401 or 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthLoginIssuesJWTCookie(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	rec := testhttp.ServeAuthz(t, router, http.MethodPost, "/api/auth/login", "", `{"email":"admin@example.com","password":"demo1234"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "tokenjoy_session_member=") {
		t.Fatalf("expected session cookie, got %q", setCookie)
	}
}

func TestDisabledMemberSessionRejected(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	memberCookie := testutil.SessionCookie(t, contract.IDMemberPure)
	disableRec := testhttp.ServeAuthz(
		t, router, http.MethodPut, "/api/org/members/status",
		testhttp.AdminCookie(t),
		`{"ids":["`+contract.IDMemberPure+`"],"status":"inactive"}`,
		nil,
	)
	if disableRec.Code != http.StatusNoContent && disableRec.Code != http.StatusOK {
		t.Fatalf("disable member: expected success, got %d body=%s", disableRec.Code, disableRec.Body.String())
	}
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/org/departments/tree", memberCookie, "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for disabled member, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDashboardCostWithoutUsagePermission(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	admin := testhttp.AdminCookie(t)
	createRec := testhttp.ServeAuthz(
		t, router, http.MethodPost, "/api/org/roles", admin,
		`{"name":"Cost Only","permissions":["p-8"]}`,
		nil,
	)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create role: expected 200, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&role); err != nil {
		t.Fatal(err)
	}
	addRec := testhttp.ServeAuthz(
		t, router, http.MethodPost, "/api/org/roles/"+role.ID+"/members", admin,
		`{"memberId":"`+contract.IDMemberPure+`"}`,
		nil,
	)
	if addRec.Code != http.StatusOK && addRec.Code != http.StatusNoContent {
		t.Fatalf("add role member: expected success, got %d body=%s", addRec.Code, addRec.Body.String())
	}

	memberCookie := testutil.SessionCookie(t, contract.IDMemberPure)
	costRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/dashboard/cost/summary", memberCookie, "", nil)
	if costRec.Code != http.StatusOK {
		t.Fatalf("expected cost 200, got %d body=%s", costRec.Code, costRec.Body.String())
	}
	usageRec := testhttp.ServeAuthz(
		t, router, http.MethodGet,
		"/api/dashboard/usage/series?granularity=day&start=2026-06-10&end=2026-06-11",
		memberCookie, "", nil,
	)
	if usageRec.Code != http.StatusForbidden {
		t.Fatalf("expected usage 403, got %d body=%s", usageRec.Code, usageRec.Body.String())
	}
}

func TestSelfApprovalWithoutKeysAdminRead(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	memberCookie := testutil.SessionCookie(t, contract.IDMemberPure)
	approvalBody := `{"type":"budget","reason":"need more","requestedBudget":500,"memberId":"` + contract.IDMemberPure + `"}`
	createRec := testhttp.ServeAuthz(t, router, http.MethodPost, "/api/keys/approvals", memberCookie, approvalBody, nil)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected approval create 200, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	platformRec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/keys/platform", memberCookie, "", nil)
	if platformRec.Code != http.StatusForbidden {
		t.Fatalf("expected platform 403, got %d body=%s", platformRec.Code, platformRec.Body.String())
	}
}
