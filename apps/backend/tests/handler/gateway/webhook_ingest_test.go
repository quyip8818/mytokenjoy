package gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func newWebhookApp(t *testing.T, mutate func(*config.Config)) *app.App {
	t.Helper()
	return testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
		testutil.WithNewAPIWebhookSecret(webhookSecret)(cfg)
		if mutate != nil {
			mutate(cfg)
		}
	})
}

func postWebhook(t *testing.T, application *app.App, logID int64) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]int64{"log_id": logID})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	return rec
}

func pendingIngestRiverJobs(t *testing.T, application *app.App) int {
	t.Helper()
	return riverfix.PendingJobCount(application.Store, jobs.KindIngest, 0)
}

func drainIngestQueue(t *testing.T, application *app.App) {
	t.Helper()
	if err := application.RunIngestOnce(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookIngestSuccess(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.PrepareIngestFixture(t, application.Store, newapisynctf.DefaultMappingOpts())

	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(92001, 99))
	rec := postWebhook(t, application, 92001)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "accepted" {
		t.Fatalf("expected status accepted, got %q", resp["status"])
	}

	// Before draining, should not be in ledger yet.
	ingested, err := testutil.HasLedgerLogID(application.Store, 92001)
	if err != nil || ingested {
		t.Fatalf("expected no ledger before worker, ingested=%v err=%v", ingested, err)
	}

	drainIngestQueue(t, application)

	ingested, err = testutil.HasLedgerLogID(application.Store, 92001)
	if err != nil || !ingested {
		t.Fatalf("expected ledger after worker, ingested=%v err=%v", ingested, err)
	}
}

func TestWebhookIngestIdempotent(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.PrepareIngestFixture(t, application.Store, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(93001, 99))
	for i := 0; i < 2; i++ {
		rec := postWebhook(t, application, 93001)
		if rec.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, rec.Code)
		}
	}
	drainIngestQueue(t, application)
	ingested, err := testutil.HasLedgerLogID(application.Store, 93001)
	if err != nil || !ingested {
		t.Fatalf("expected single ledger row, ingested=%v err=%v", ingested, err)
	}
}

func TestWebhookIngestWritesLedgerFields(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.PrepareIngestFixture(t, application.Store, newapisynctf.DefaultMappingOpts())

	const input = "webhook preview"
	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(94002, 99))
	rec := postWebhook(t, application, 94002)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	drainIngestQueue(t, application)

	exists, err := testutil.HasLedgerLogID(application.Store, 94002)
	if err != nil || !exists {
		t.Fatalf("expected ledger entry, exists=%v err=%v", exists, err)
	}
}

func TestWebhookLogNotFoundEnqueuesRiverJob(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	rec := postWebhook(t, application, 99999)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// Should have enqueued a River ingest job even though the log doesn't exist.
	if n := pendingIngestRiverJobs(t, application); n == 0 {
		t.Fatal("expected pending river ingest job")
	}
}

func TestIngestMetricsEndpoint(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.PrepareIngestFixture(t, application.Store, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(98100, 99))

	rec := postWebhook(t, application, 98100)
	if rec.Code != http.StatusOK {
		t.Fatalf("webhook expected 200, got %d", rec.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/api/internal/metrics/ingest", nil)
	metricsReq.Header.Set("X-Webhook-Secret", webhookSecret)
	metricsRec := httptest.NewRecorder()
	application.Router.ServeHTTP(metricsRec, metricsReq)
	if metricsRec.Code != http.StatusOK {
		t.Fatalf("metrics expected 200, got %d body=%s", metricsRec.Code, metricsRec.Body.String())
	}
	var snap map[string]any
	if err := json.NewDecoder(metricsRec.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if _, ok := snap["ingest_webhook_accepted_total"]; !ok {
		t.Fatalf("missing ingest_webhook_accepted_total in %v", snap)
	}
	if notify, _ := snap["ingest_webhook_accepted_total"].(float64); notify < 1 {
		t.Fatalf("expected webhook accepted counter >= 1, got %v", snap["ingest_webhook_accepted_total"])
	}
}

func TestIngestMetricsDisabledReturns404(t *testing.T) {
	t.Parallel()
	application := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithNewAPIWebhookSecret(webhookSecret)(cfg)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/internal/metrics/ingest", nil)
	rec := httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when ingest disabled, got %d", rec.Code)
	}
}
