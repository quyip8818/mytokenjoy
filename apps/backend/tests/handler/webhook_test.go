package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
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
	tokenID := int64(99)
	if err := app.Store.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      testutil.StrPtr(seed.IDMember1),
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}

	beforeConsumed := dept3Consumed(t, app.Store.Budget().Tree())

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

	afterConsumed := dept3Consumed(t, app.Store.Budget().Tree())
	if afterConsumed <= beforeConsumed {
		t.Fatalf("expected consumed rollup, before=%v after=%v", beforeConsumed, afterConsumed)
	}
}

func dept3Consumed(t *testing.T, tree []types.BudgetNode) float64 {
	t.Helper()
	consumed, ok := findDept3Consumed(tree)
	if !ok {
		t.Fatal("dept-3 not found in budget tree")
	}
	return consumed
}

func findDept3Consumed(tree []types.BudgetNode) (float64, bool) {
	var walk func([]types.BudgetNode) (float64, bool)
	walk = func(nodes []types.BudgetNode) (float64, bool) {
		for _, node := range nodes {
			if node.ID == seed.IDDept3 {
				return node.Consumed, true
			}
			if len(node.Children) > 0 {
				if v, ok := walk(node.Children); ok {
					return v, true
				}
			}
		}
		return 0, false
	}
	return walk(tree)
}
