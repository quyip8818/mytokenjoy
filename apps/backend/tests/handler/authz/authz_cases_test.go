package authz_test

import (
	"fmt"
	"net/http"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

const validDept3Budget = 20000

type authzCase struct {
	name       string
	method     string
	path       string
	body       string
	cookie     string
	headers    map[string]string
	wantStatus int
}

func authzWriteCases(t *testing.T) []authzCase {
	t.Helper()
	memberCreateBody := fmt.Sprintf(
		`{"name":"X","phone":"13900000001","email":"x@example.com","departmentId":%q}`,
		seed.IDDept3,
	)
	platformKeyBody := fmt.Sprintf(
		`{"name":"k","memberId":%q,"quota":100,"modelWhitelist":["gpt-4o"]}`,
		seed.IDMember1,
	)
	budgetUpdateBody := fmt.Sprintf(`{"budget":%d}`, validDept3Budget)
	deptCreateBody := fmt.Sprintf(`{"name":"Auth Test","parentId":%q}`, seed.IDDept2)
	adminCookie := testhttp.AdminCookie(t)
	pureCookie := testutil.SessionCookie(t, seed.IDMemberPure)

	return []authzCase{
		{name: "department create unauthorized", method: http.MethodPost, path: "/api/org/departments", body: deptCreateBody, wantStatus: http.StatusUnauthorized},
		{name: "department create forbidden", method: http.MethodPost, path: "/api/org/departments", body: deptCreateBody, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "department create allowed", method: http.MethodPost, path: "/api/org/departments", body: deptCreateBody, cookie: adminCookie, wantStatus: http.StatusOK},
		{name: "budget update unauthorized", method: http.MethodPut, path: "/api/budget/departments/" + seed.IDDept3, body: budgetUpdateBody, wantStatus: http.StatusUnauthorized},
		{name: "budget update forbidden", method: http.MethodPut, path: "/api/budget/departments/" + seed.IDDept3, body: `{"budget":1000}`, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "budget update allowed", method: http.MethodPut, path: "/api/budget/departments/" + seed.IDDept3, body: budgetUpdateBody, cookie: adminCookie, wantStatus: http.StatusOK},
		{name: "keys platform create unauthorized", method: http.MethodPost, path: "/api/keys/platform", body: platformKeyBody, wantStatus: http.StatusUnauthorized},
		{name: "keys platform create forbidden", method: http.MethodPost, path: "/api/keys/platform", body: platformKeyBody, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "model create forbidden", method: http.MethodPost, path: "/api/models", body: `{"name":"test-model","displayName":"Test","baseUrl":"http://x","inputPrice":1,"outputPrice":2}`, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "org member create forbidden", method: http.MethodPost, path: "/api/org/members", body: memberCreateBody, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "audit settings forbidden", method: http.MethodPut, path: "/api/audit/settings", body: `{"retentionDays":30}`, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "datasource update forbidden", method: http.MethodPut, path: "/api/org/data-source", body: `{"platform":"feishu","appId":"a","appSecret":"b"}`, cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "dashboard forbidden without permission", method: http.MethodGet, path: "/api/dashboard/cost/summary", cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "billing wallet forbidden", method: http.MethodGet, path: "/api/billing/wallet", cookie: pureCookie, wantStatus: http.StatusForbidden},
		{name: "sync trigger unauthorized", method: http.MethodPost, path: "/api/org/sync/trigger", wantStatus: http.StatusUnauthorized},
	}
}

func prodGetForbiddenCases(t *testing.T) []authzCase {
	t.Helper()
	memberCookie := testutil.SessionCookie(t, seed.IDMember1)
	return []authzCase{
		{name: "org departments tree", method: http.MethodGet, path: "/api/org/departments/tree", cookie: memberCookie, wantStatus: http.StatusForbidden},
		{name: "budget tree", method: http.MethodGet, path: "/api/budget/tree", cookie: memberCookie, wantStatus: http.StatusForbidden},
		{name: "keys provider", method: http.MethodGet, path: "/api/keys/provider", cookie: memberCookie, wantStatus: http.StatusForbidden},
		{name: "models list", method: http.MethodGet, path: "/api/models", cookie: memberCookie, wantStatus: http.StatusForbidden},
		{name: "audit settings", method: http.MethodGet, path: "/api/audit/settings", cookie: memberCookie, wantStatus: http.StatusForbidden},
		{name: "dashboard cost summary", method: http.MethodGet, path: "/api/dashboard/cost/summary", cookie: memberCookie, wantStatus: http.StatusForbidden},
	}
}
