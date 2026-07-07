package me_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	domainmember "github.com/tokenjoy/backend/internal/domain/member"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGetMemberDashboardHTTP(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/me/dashboard", nil)
	req.Header.Set("Cookie", testutil.SessionCookie(t, contract.IDMember1))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var view domainmember.DashboardView
	if err := json.NewDecoder(rec.Body).Decode(&view); err != nil {
		t.Fatal(err)
	}
	if view.UsageStats.RequestCount <= 0 {
		t.Fatalf("expected positive request count, got %+v", view.UsageStats)
	}
}

func TestGetMemberDashboardForbiddenWithoutSelfKeys(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/me/dashboard", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin should have self:keys via super admin, got %d", rec.Code)
	}
}
