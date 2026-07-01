package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

const webhookSecret = "test-secret"

func TestWebhookUnauthorized(t *testing.T) {
	app := newTestApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	router := app.Router

	body, _ := json.Marshal(newapi.WebhookLogPayload{ID: 1, TokenID: 99})
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
	app := newTestApp(t, func(cfg *config.Config) {
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
	app := newTestApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	beforeBuckets := testutil.UsageBucketCount(app.Store)
	testutil.UpsertRelayMapping(t, app.Store, testutil.DefaultRelayMappingOpts())

	ctx := testutil.Ctx()
	budgetTree, err := app.Store.Budget().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeConsumed := testutil.Dept3Consumed(t, budgetTree)
	beforeUsed := platformKeyUsed(t, app.Store, seed.IDPlatformKey1)

	payload := newapi.WebhookLogPayload{
		ID: 2001, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1,
	}
	body, _ := json.Marshal(payload)
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

	budgetTree, err = app.Store.Budget().Tree(ctx)
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
	app := newTestApp(t, func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = webhookSecret
	})
	beforeBuckets := testutil.UsageBucketCount(app.Store)
	testutil.UpsertRelayMapping(t, app.Store, testutil.DefaultRelayMappingOpts())
	payload := newapi.WebhookLogPayload{ID: 3001, TokenID: 99, Quota: 500000, Model: "gpt-4o", CreatedAt: 1}
	body, _ := json.Marshal(payload)
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
