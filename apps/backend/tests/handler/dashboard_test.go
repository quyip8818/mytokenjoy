package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUsageSeriesMinuteDemoFallbackWithoutNewAPI(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T10:00:00%2B08:00&end=2026-06-10T11:00:00%2B08:00", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceBuckets || !resp.Approximate {
		t.Fatalf("expected demo bucket fallback, got %+v", resp)
	}
}

func TestUsageSeriesMinuteUnavailableInProdProfile(t *testing.T) {
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		cfg.Profile = config.ProfileProd
	})
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T10:00:00%2B08:00&end=2026-06-10T11:00:00%2B08:00", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["retryAfter"].(float64) != float64(types.UsageMinuteRetryAfterSecs) {
		t.Fatalf("expected retryAfter=%d, got %+v", types.UsageMinuteRetryAfterSecs, body["retryAfter"])
	}
}

func TestCostDailyInvalidGranularity(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/daily?period=current_month&granularity=minute", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUsageSeriesWindowTooLarge(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=day&start=2024-01-01&end=2026-01-01", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDashboardUnauthorizedWithoutSession(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/summary", nil)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUsageSeriesGroupByDepartmentHTTP(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	testutil.SeedUsageBucket(t, app.Store, testutil.UsageBucketOpts{CostCNY: 4})
	testutil.SeedUsageBucket(t, app.Store, testutil.UsageBucketOpts{
		BucketStart:  time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
		DepartmentID: seed.IDDept4, MemberID: "m-4", CostCNY: 6,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=day&start=2026-06-10&end=2026-06-11&groupBy=department", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceBuckets || len(resp.Points) != 2 {
		t.Fatalf("expected two department points from buckets, got %+v", resp)
	}
}

func TestUsageSeriesMinuteSuccessMetaHTTP(t *testing.T) {
	var serverURL string
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		now := time.Now().UTC()
		server := testutil.StartNewAPIMockServer(t, []newapi.LogEntry{
			testutil.SampleMappedLog(42, now.Add(-10*time.Minute)),
		})
		serverURL = server.URL
		cfg.NewAPIEnabled = true
		cfg.NewAPIBaseURL = serverURL
		cfg.NewAPIAdminToken = "test-token"
	})
	testutil.UpsertRelayMapping(t, app.Store, testutil.RelayMappingOpts{
		PlatformKeyID: "plk-minute-test", NewAPITokenID: 42,
	})
	start := time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339)
	end := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start="+start+"&end="+end+"&groupBy=none", nil)
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLogs || !resp.Approximate || resp.MappingAsOf != types.UsageMappingAsOfQueryTime {
		t.Fatalf("unexpected minute response meta: %+v", resp)
	}
}

func TestDashboardEndpointsAreReadOnly(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	testutil.SeedUsageBucket(t, app.Store, testutil.DefaultUsageBucketOpts())
	beforeBuckets := testutil.UsageBucketCount(app.Store)
	paths := []string{
		"/api/dashboard/cost/summary",
		"/api/dashboard/cost/departments",
		"/api/dashboard/cost/departments/" + seed.IDDept3 + "/members",
		"/api/dashboard/cost/daily",
		"/api/dashboard/cost/top?limit=5",
		"/api/dashboard/usage/models",
		"/api/dashboard/usage/teams",
		"/api/dashboard/usage/series?granularity=day&start=2026-06-10&end=2026-06-11",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Cookie", sessionCookie)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s expected 200, got %d", path, rec.Code)
		}
	}
	testutil.AssertUsageBucketCount(t, app.Store, beforeBuckets)
}
