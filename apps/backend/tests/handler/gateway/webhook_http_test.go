package gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

const webhookSecret = "test-secret"

func TestWebhookUnauthorized(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
		testutil.WithNewAPIWebhookSecret(webhookSecret)(cfg)
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
	t.Parallel()
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithIngestEnabled(true)(cfg)
		testutil.WithNewAPIWebhookSecret(webhookSecret)(cfg)
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
