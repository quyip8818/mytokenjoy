package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDashboardDefaultApp(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	testutil.ApplyDemoRuntime(t, app.Store, app.Config)
	adminCookie := testhttp.AdminCookie(t)

	t.Run("cost daily invalid granularity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/daily?period=current_month&granularity=minute", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("unauthorized without session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/summary", nil)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("endpoints are read only", func(t *testing.T) {
		testutil.SeedUsageBucket(t, app.Store, testutil.DefaultUsageBucketOpts())
		beforeBuckets := testutil.UsageBucketCount(app.Store)
		paths := []string{
			"/api/dashboard/cost/summary",
			"/api/dashboard/cost/departments",
			"/api/dashboard/cost/departments/" + contract.IDDept3.String() + "/members",
			"/api/dashboard/cost/daily",
			"/api/dashboard/cost/top?limit=5",
			"/api/dashboard/usage/models",
			"/api/dashboard/usage/teams",
		}
		for _, path := range paths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Cookie", adminCookie)
			rec := httptest.NewRecorder()
			app.Router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("path %s expected 200, got %d", path, rec.Code)
			}
		}
		testutil.AssertUsageBucketCount(t, app.Store, beforeBuckets)
	})
}
