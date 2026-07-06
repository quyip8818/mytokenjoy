package gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

const webhookSecret = "test-secret"

func TestWebhookUnauthorized(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	router := app.Router

	body, _ := json.Marshal(map[string]int64{"log_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without secret, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", "wrong")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong secret, got %d", rec.Code)
	}
}

func TestWebhookInvalidPayload(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestWebhookIngestSuccess(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	beforeBuckets := testutil.UsageBucketCount(app.Store)
	relayfix.UpsertMapping(t, app.Store, relayfix.DefaultMappingOpts())

	ctx := testutil.Ctx()
	budgetTree, err := common.LoadBudgetTree(ctx, app.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	beforeConsumed := testutil.Dept3Consumed(t, budgetTree)
	beforeUsed := platformKeyUsed(t, app.Store, seed.IDPlatformKey1)

	testutil.SeedConsumeLog(t, app.Store, testutil.DefaultConsumeLog(2001, 99))
	body, _ := json.Marshal(map[string]int64{"log_id": 2001})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", resp["status"])
	}

	budgetTree, err = common.LoadBudgetTree(ctx, app.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	afterConsumed := testutil.Dept3Consumed(t, budgetTree)
	if afterConsumed <= beforeConsumed {
		t.Fatalf("expected consumed rollup, before=%v after=%v", beforeConsumed, afterConsumed)
	}
	afterUsed := platformKeyUsed(t, app.Store, seed.IDPlatformKey1)
	if afterUsed <= beforeUsed {
		t.Fatalf("expected platform key used increase, before=%v after=%v", beforeUsed, afterUsed)
	}
	testutil.AssertUsageBucketCount(t, app.Store, beforeBuckets+1)
}

func TestWebhookIngestIdempotent(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	beforeBuckets := testutil.UsageBucketCount(app.Store)
	relayfix.UpsertMapping(t, app.Store, relayfix.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, app.Store, testutil.DefaultConsumeLog(3001, 99))
	body, _ := json.Marshal(map[string]int64{"log_id": 3001})
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Secret", webhookSecret)
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200, got %d", i+1, rec.Code)
		}
	}
	testutil.AssertUsageBucketCount(t, app.Store, beforeBuckets+1)
}

func TestWebhookIngestWritesLedgerFields(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	relayfix.UpsertMapping(t, app.Store, relayfix.DefaultMappingOpts())

	const input = "webhook preview"
	testutil.SeedConsumeLog(t, app.Store, store.RawConsumeLog{
		ID: 4002, TokenID: 99, Quota: 500000, ModelName: "gpt-4o", CreatedAt: 1717200000,
		PromptTokens: 88, CompletionTokens: 22, UseTime: 99, Content: input,
	})
	body, _ := json.Marshal(map[string]int64{"log_id": 4002})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	exists, err := testutil.HasLedgerLogID(app.Store, 4002)
	if err != nil || !exists {
		t.Fatalf("expected ledger entry, exists=%v err=%v", exists, err)
	}

	entries, _, err := app.Store.Ledger().ListCallSettledPage(testutil.Ctx(), store.LedgerCallFilter{Page: 1, PageSize: 10000})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, entry := range entries {
		if entry.IdempotencyKey == "newapi:4002" {
			found = true
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
		t.Fatal("ledger entry 4002 not found")
	}
}

func TestWebhookLogNotFoundReturns503(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	body, _ := json.Marshal(map[string]int64{"log_id": 99999})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestWebhookMappingMissingAcceptsAndRecordsFailure(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	testutil.SeedConsumeLog(t, app.Store, testutil.DefaultConsumeLog(8001, 55))
	body, _ := json.Marshal(map[string]int64{"log_id": 8001})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	testutil.AssertIngestFailure(t, app.Store, 8001, types.SourceWebhook)
	ingested, err := testutil.HasLedgerLogID(app.Store, 8001)
	if err != nil || ingested {
		t.Fatalf("expected no ledger entry, ingested=%v err=%v", ingested, err)
	}
}

func TestIngestMetricsEndpoint(t *testing.T) {
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	relayfix.UpsertMapping(t, app.Store, relayfix.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, app.Store, testutil.DefaultConsumeLog(8100, 99))

	body, _ := json.Marshal(map[string]int64{"log_id": 8100})
	req := httptest.NewRequest(http.MethodPost, "/api/internal/webhooks/newapi-log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", webhookSecret)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("webhook expected 200, got %d", rec.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/api/internal/metrics/ingest", nil)
	metricsRec := httptest.NewRecorder()
	app.Router.ServeHTTP(metricsRec, metricsReq)
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
	app := testhttp.NewApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/internal/metrics/ingest", nil)
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when ingest disabled, got %d", rec.Code)
	}
}

func platformKeyUsed(t *testing.T, st store.Store, keyID string) float64 {
	t.Helper()
	keys, err := st.Keys().PlatformKeys(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == keyID {
			return key.Used
		}
	}
	t.Fatalf("platform key %s not found", keyID)
	return 0
}
