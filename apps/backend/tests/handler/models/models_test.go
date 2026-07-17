package models_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
)

func TestRoutingUpdateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"allowedModelIds":[100]}`)
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
	if len(rule.AllowedModelIDs) != 1 || rule.AllowedModelIDs[0] != contract.IDModel1 {
		t.Fatalf("expected allowedModelIds [%s], got %v", contract.IDModel1, rule.AllowedModelIDs)
	}
	if len(rule.AllowedModels) != 1 || rule.AllowedModels[0].Type != "gpt-4o" {
		t.Fatalf("expected enriched gpt-4o, got %v", rule.AllowedModels)
	}
}

func TestModelCreateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"type":"custom-model","name":"Custom","baseUrl":"http://llm.test","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
	req := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var model types.ModelInfo
	if err := json.NewDecoder(rec.Body).Decode(&model); err != nil {
		t.Fatal(err)
	}
	if model.Type != "custom-model" {
		t.Fatalf("expected type custom-model, got %q", model.Type)
	}
	if model.Endpoint == nil || *model.Endpoint != "http://llm.test" {
		t.Fatalf("expected endpoint http://llm.test, got %+v", model.Endpoint)
	}
}

func TestModelListHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var models []types.ModelInfo
	if err := json.NewDecoder(rec.Body).Decode(&models); err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("expected seeded models")
	}
	found := false
	for _, model := range models {
		if model.Provider != "" && model.Type != "" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected models with type and provider, got %+v", models[0])
	}
}

func TestModelUpdateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	createBody := []byte(`{"type":"edit-model","name":"Edit","baseUrl":"http://llm.old","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Cookie", testhttp.AdminCookie(t))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create expected 200, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created types.ModelInfo
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	updateBody := []byte(`{"name":"Edited","description":"test desc","endpoint":"http://llm.new"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/models/"+created.ID.String(), bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Cookie", testhttp.AdminCookie(t))
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}
	var updated types.ModelInfo
	if err := json.NewDecoder(updateRec.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Edited" || updated.Description != "test desc" {
		t.Fatalf("unexpected update fields: %+v", updated)
	}
	if updated.Endpoint == nil || *updated.Endpoint != "http://llm.new" {
		t.Fatalf("expected endpoint http://llm.new, got %+v", updated.Endpoint)
	}
}

func TestModelToggleHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	createBody := []byte(`{"type":"toggle-model","name":"Toggle","baseUrl":"http://llm.toggle","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Cookie", testhttp.AdminCookie(t))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create expected 200, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created types.ModelInfo
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	body := []byte(`{"enabled":false}`)
	req := httptest.NewRequest(http.MethodPut, "/api/models/"+created.ID.String()+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
