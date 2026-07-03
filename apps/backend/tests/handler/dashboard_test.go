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

func TestUsageSeriesMinuteFromLedgerDemo(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T08:00:00Z&end=2026-06-10T10:00:00Z", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLedger || resp.Approximate {
		t.Fatalf("expected ledger minute series, got %+v", resp)
	}
	if len(resp.Points) == 0 {
		t.Fatalf("expected seeded ledger points in demo window, got %+v", resp)
	}
}

func TestUsageSeriesMinuteFromLedgerProdProfile(t *testing.T) {
	app := testutil.NewTestApp(t, func(cfg *config.Config) {
		cfg.Profile = config.ProfileProd
	})
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T08:00:00Z&end=2026-06-10T10:00:00Z", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLedger {
		t.Fatalf("expected ledger source, got %+v", resp)
	}
}

func TestCostDailyInvalidGranularity(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/daily?period=current_month&granularity=minute", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUsageSeriesWindowTooLarge(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=day&start=2024-01-01&end=2026-01-01", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
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
	req.Header.Set("Cookie", adminSessionCookie(t))
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
	app := testutil.NewTestApp(t, nil)
	testutil.UpsertRelayMapping(t, app.Store, testutil.RelayMappingOpts{
		PlatformKeyID: "plk-minute-test", NewAPITokenID: 42,
	})
	ingest := testutil.NewIngestService(t, testutil.TestConfig(), app.Store)
	occurredAt := time.Date(2026, 6, 10, 9, 3, 0, 0, time.UTC)
	if err := ingest.Ingest(testutil.Ctx(), newapi.WebhookLogPayload{
		ID: 88001, TokenID: 42, Quota: 500000, Model: "gpt-4o", CreatedAt: occurredAt.Unix(),
		PromptTokens: 100, CompletionTokens: 50, UseTime: 200,
	}, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T09:00:00Z&end=2026-06-10T10:00:00Z&groupBy=none", nil)
	req.Header.Set("Cookie", adminSessionCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.UsageSeriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLedger || resp.Approximate || resp.MappingAsOf != types.UsageMappingAsOfIngestTime {
		t.Fatalf("unexpected minute response meta: %+v", resp)
	}
	if len(resp.Points) == 0 {
		t.Fatalf("expected minute points from ledger, got %+v", resp)
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
		req.Header.Set("Cookie", adminSessionCookie(t))
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s expected 200, got %d", path, rec.Code)
		}
	}
	testutil.AssertUsageBucketCount(t, app.Store, beforeBuckets)
}
