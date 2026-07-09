package dashboard_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDashboardDefaultApp(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	adminCookie := testhttp.AdminCookie(t)

	t.Run("usage series minute from ledger demo", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T08:00:00Z&end=2026-06-10T10:00:00Z", nil)
		req.Header.Set("Cookie", adminCookie)
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
	})

	t.Run("cost daily invalid granularity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/cost/daily?period=current_month&granularity=minute", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("usage series window too large", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=day&start=2024-01-01&end=2026-01-01", nil)
		req.Header.Set("Cookie", adminCookie)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
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

	t.Run("usage series group by department", func(t *testing.T) {
		testutil.SeedUsageBucket(t, app.Store, testutil.UsageBucketOpts{Cost: 4})
		testutil.SeedUsageBucket(t, app.Store, testutil.UsageBucketOpts{
			BucketStart:  time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept4, MemberID: "m-4", Cost: 6,
		})
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=day&start=2026-06-10&end=2026-06-11&groupBy=department", nil)
		req.Header.Set("Cookie", adminCookie)
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
	})

	t.Run("endpoints are read only", func(t *testing.T) {
		testutil.SeedUsageBucket(t, app.Store, testutil.DefaultUsageBucketOpts())
		beforeBuckets := testutil.UsageBucketCount(app.Store)
		paths := []string{
			"/api/dashboard/cost/summary",
			"/api/dashboard/cost/departments",
			"/api/dashboard/cost/departments/" + contract.IDDept3 + "/members",
			"/api/dashboard/cost/daily",
			"/api/dashboard/cost/top?limit=5",
			"/api/dashboard/usage/models",
			"/api/dashboard/usage/teams",
			"/api/dashboard/usage/series?granularity=day&start=2026-06-10&end=2026-06-11",
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

func TestUsageSeriesMinuteFromLedger(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T08:00:00Z&end=2026-06-10T10:00:00Z", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
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

func TestUsageSeriesMinuteSuccessMetaHTTP(t *testing.T) {
	t.Parallel()
	app := newIngestDashboardApp(t)
	ctx := testutil.Ctx()
	memberID := contract.IDMember1
	if err := app.Store.Keys().SetPlatformKeys(ctx, []types.PlatformKey{{
		ID:        "plk-minute-test",
		Name:      "Minute Test Key",
		KeyPrefix: "sk-minute",
		MemberID:  &memberID,
		Status:    "active",
		CreatedAt: "2026-06-19",
	}}); err != nil {
		t.Fatal(err)
	}
	relayfix.UpsertMapping(t, app.Store, relayfix.MappingOpts{
		PlatformKeyID: "plk-minute-test", NewAPITokenID: 42,
	})
	ingest := testutil.NewIngestService(t, testutil.TestConfig(testutil.WithIngestEnabled(true)), app.Store)
	occurredAt := time.Date(2026, 6, 10, 9, 3, 0, 0, time.UTC)
	testutil.SeedConsumeLog(t, app.Store, store.RawConsumeLog{
		ID: 88001, TokenID: 42, Quota: 500000, ModelName: "gpt-4o", CreatedAt: occurredAt.Unix(),
		PromptTokens: 100, CompletionTokens: 50, UseTime: 200,
	})
	if err := ingest.IngestByLogID(testutil.Ctx(), 88001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/usage/series?granularity=minute&start=2026-06-10T09:00:00Z&end=2026-06-10T10:00:00Z&groupBy=none", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
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

func newIngestDashboardApp(t *testing.T) *app.App {
	t.Helper()
	return testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
	})
}
