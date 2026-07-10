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
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
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

func drainIngestQueue(t *testing.T, application *app.App) {
	t.Helper()
	application.Worker.RunOnce(testutil.Ctx())
}

func TestWebhookIngestSuccess(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	beforeBuckets := testutil.UsageBucketCount(application.Store)
	newapisynctf.UpsertMapping(t, application.Store, newapisynctf.DefaultMappingOpts())

	beforeConsumed := testutil.Dept3SnapshotConsumed(t, application.Store)
	beforeUsed := testutil.PlatformKeySnapshotUsed(t, application.Store, contract.IDPlatformKey1)

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
	testutil.AssertIngestJob(t, application.Store, 92001, types.SourceWebhook)
	ingested, err := testutil.HasLedgerLogID(application.Store, 92001)
	if err != nil || ingested {
		t.Fatalf("expected no ledger before worker, ingested=%v err=%v", ingested, err)
	}

	drainIngestQueue(t, application)

	afterConsumed := testutil.Dept3SnapshotConsumed(t, application.Store)
	if afterConsumed <= beforeConsumed {
		t.Fatalf("expected consumed rollup, before=%v after=%v", beforeConsumed, afterConsumed)
	}
	afterUsed := testutil.PlatformKeySnapshotUsed(t, application.Store, contract.IDPlatformKey1)
	if afterUsed <= beforeUsed {
		t.Fatalf("expected platform key used increase, before=%v after=%v", beforeUsed, afterUsed)
	}
	testutil.AssertUsageBucketCount(t, application.Store, beforeBuckets+1)
	if n := testutil.PendingIngestJobCount(t, application.Store); n != 0 {
		t.Fatalf("expected queue drained, pending=%d", n)
	}
}

func TestWebhookIngestIdempotent(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	beforeBuckets := testutil.UsageBucketCount(application.Store)
	newapisynctf.UpsertMapping(t, application.Store, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(93001, 99))
	for i := 0; i < 2; i++ {
		rec := postWebhook(t, application, 93001)
		if rec.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, rec.Code)
		}
	}
	drainIngestQueue(t, application)
	testutil.AssertUsageBucketCount(t, application.Store, beforeBuckets+1)
}

func TestWebhookIngestWritesLedgerFields(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.UpsertMapping(t, application.Store, newapisynctf.DefaultMappingOpts())

	const input = "webhook preview"
	testutil.SeedConsumeLog(t, application.Store, store.RawConsumeLog{
		ID: 94002, TokenID: 99, Quota: 500000, ModelName: "gpt-4o", CreatedAt: 1717200000,
		PromptTokens: 88, CompletionTokens: 22, UseTime: 99, Content: input,
	})
	rec := postWebhook(t, application, 94002)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	drainIngestQueue(t, application)

	exists, err := testutil.HasLedgerLogID(application.Store, 94002)
	if err != nil || !exists {
		t.Fatalf("expected ledger entry, exists=%v err=%v", exists, err)
	}

	entries, _, err := application.Store.Ledger().ListCallSettledPage(testutil.Ctx(), store.LedgerCallFilter{Page: 1, PageSize: 10000})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, entry := range entries {
		if entry.IdempotencyKey == "newapi:94002" {
			found = true
			if entry.Source != types.SourceWebhook {
				t.Fatalf("expected source webhook, got %q", entry.Source)
			}
			if entry.InputTokens != 88 || entry.OutputTokens != 22 {
				t.Fatalf("unexpected token counts %d/%d", entry.InputTokens, entry.OutputTokens)
			}
			if entry.CallDetail.PreviewSnippet != input {
				t.Fatalf("unexpected snippet %q", entry.CallDetail.PreviewSnippet)
			}
			break
		}
	}
	if !found {
		t.Fatal("ledger entry 94002 not found")
	}
}

func TestWebhookLogNotFoundStillAccepted(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	rec := postWebhook(t, application, 99999)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	testutil.AssertIngestJob(t, application.Store, 99999, types.SourceWebhook)
	drainIngestQueue(t, application)
	f := testutil.AssertIngestJob(t, application.Store, 99999, types.SourceWebhook)
	if f.Status != store.IngestJobStatusPending {
		t.Fatalf("expected pending retry after log-not-found, got %q", f.Status)
	}
	if f.Attempts < 1 {
		t.Fatalf("expected attempts incremented, got %d", f.Attempts)
	}
}

func TestWebhookMappingMissingAcceptedThenWorkerRecords(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	testutil.SeedConsumeLog(t, application.Store, testutil.DefaultConsumeLog(98001, 55))
	rec := postWebhook(t, application, 98001)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	testutil.AssertIngestJob(t, application.Store, 98001, types.SourceWebhook)
	drainIngestQueue(t, application)
	f := testutil.AssertIngestJob(t, application.Store, 98001, types.SourceWebhook)
	if f.Status != store.IngestJobStatusPending && f.Status != store.IngestJobStatusDead {
		t.Fatalf("expected pending or dead after mapping miss, got %q", f.Status)
	}
	ingested, err := testutil.HasLedgerLogID(application.Store, 98001)
	if err != nil || ingested {
		t.Fatalf("expected no ledger entry, ingested=%v err=%v", ingested, err)
	}
}

func TestIngestMetricsEndpoint(t *testing.T) {
	t.Parallel()
	application := newWebhookApp(t, nil)
	newapisynctf.UpsertMapping(t, application.Store, newapisynctf.DefaultMappingOpts())
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
	if _, ok := snap["ingest_notify_total"]; !ok {
		t.Fatalf("missing ingest_notify_total in %v", snap)
	}
	if notify, _ := snap["ingest_notify_total"].(float64); notify < 1 {
		t.Fatalf("expected notify counter >= 1, got %v", snap["ingest_notify_total"])
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
