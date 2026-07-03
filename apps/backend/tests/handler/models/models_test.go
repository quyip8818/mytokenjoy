package models_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestRoutingUpdateHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	body := []byte(`{"allowedModels":["gpt-4o"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/models/routing/dept-3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var rule types.RoutingRule
	if err := json.NewDecoder(rec.Body).Decode(&rule); err != nil {
		t.Fatal(err)
	}
	if len(rule.AllowedModels) != 1 || rule.AllowedModels[0] != "gpt-4o" {
		t.Fatalf("expected allowedModels [gpt-4o], got %v", rule.AllowedModels)
	}
}

func TestModelCreateHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	body := []byte(`{"name":"custom-model","displayName":"Custom","baseUrl":"http://llm.test","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
	req := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestModelToggleHTTP(t *testing.T) {
	router := testhttp.NewRouter(t)
	body := []byte(`{"enabled":false}`)
	req := httptest.NewRequest(http.MethodPut, "/api/models/model-1/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
