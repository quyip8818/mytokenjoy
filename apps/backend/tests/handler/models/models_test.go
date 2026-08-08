package models_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func TestModelsLocal(t *testing.T) {
	localApp := testhttp.NewApp(t, func(cfg *config.Config) { cfg.SupportSaas = false })
	router := localApp.Router
	cookie := testhttp.AdminCookie(t)

	t.Run("RoutingUpdate", func(t *testing.T) {
		t.Parallel()
		body := []byte(fmt.Sprintf(`{"allowedModelIds":["%s"]}`, contract.IDModel1))
		req := httptest.NewRequest(http.MethodPut, "/api/models/routing/"+contract.IDDept3.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
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
	})

	t.Run("ModelCreate", func(t *testing.T) {
		t.Parallel()
		body := []byte(`{"type":"custom-model","name":"Custom","baseUrl":"http://llm.test","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
		req := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
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
	})

	t.Run("ModelList", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
		req.Header.Set("Cookie", cookie)
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
	})

	t.Run("ModelUpdate", func(t *testing.T) {
		t.Parallel()
		createBody := []byte(`{"type":"edit-model","name":"Edit","baseUrl":"http://llm.old","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
		createReq := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Cookie", cookie)
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
		updateReq.Header.Set("Cookie", cookie)
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
	})

	t.Run("ModelToggle", func(t *testing.T) {
		t.Parallel()
		createBody := []byte(`{"type":"toggle-model","name":"Toggle","baseUrl":"http://llm.toggle","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
		createReq := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Cookie", cookie)
		createRec := httptest.NewRecorder()
		router.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusOK {
			t.Fatalf("create expected 200, got %d body=%s", createRec.Code, createRec.Body.String())
		}
		var created types.ModelInfo
		if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
			t.Fatal(err)
		}

		body := []byte(`{"deprecated":true}`)
		req := httptest.NewRequest(http.MethodPut, "/api/models/"+created.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestModelsSaaS(t *testing.T) {
	router := testhttp.NewRouter(t) // SaaS mode (default)
	cookie := testhttp.AdminCookie(t)

	t.Run("CreateModelRejected", func(t *testing.T) {
		t.Parallel()
		body := []byte(`{"type":"saas-rejected","name":"Rejected","baseUrl":"http://llm.test","apiKey":"secret","inputPrice":1,"outputPrice":2}`)
		req := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 in SaaS mode, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ListSanitizesProvider", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
		req.Header.Set("Cookie", cookie)
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
		for _, m := range models {
			if m.Source == "platform" || m.Source == "seed" {
				if m.Provider != modelcatalog.DisplayProvider {
					t.Fatalf("expected provider %q for %s model %q, got %q",
						modelcatalog.DisplayProvider, m.Source, m.Type, m.Provider)
				}
			}
		}
	})
}
