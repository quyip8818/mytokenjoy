package platform_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestPlatformSetAndListGlobalPricing(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	cookie := saas.LoginPlatform(t, router)

	// Set global pricing
	body, _ := json.Marshal(map[string]any{
		"modelType":   "deepseek-chat",
		"inputPrice":  2.0,
		"outputPrice": 8.0,
		"note":        "initial pricing",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("set global pricing: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// List global pricing
	req = httptest.NewRequest(http.MethodGet, "/api/platform/pricing", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list global pricing: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var prices []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&prices); err != nil {
		t.Fatal(err)
	}
	if len(prices) == 0 {
		t.Fatal("expected at least one pricing row")
	}
	found := false
	for _, p := range prices {
		if p["modelType"] == "deepseek-chat" {
			found = true
			if p["inputPrice"].(float64) != 2.0 {
				t.Errorf("inputPrice = %v, want 2.0", p["inputPrice"])
			}
			if p["outputPrice"].(float64) != 8.0 {
				t.Errorf("outputPrice = %v, want 8.0", p["outputPrice"])
			}
		}
	}
	if !found {
		t.Fatal("deepseek-chat not found in pricing list")
	}
}

func TestPlatformGlobalPriceHistory(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	cookie := saas.LoginPlatform(t, router)

	// Set two prices for same model (append-only timeline)
	for _, price := range []float64{3.0, 1.5} {
		body, _ := json.Marshal(map[string]any{
			"modelType":   "gpt-4o",
			"inputPrice":  price,
			"outputPrice": price * 4,
			"note":        "price change",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("set pricing: expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	}

	// Get history
	req := httptest.NewRequest(http.MethodGet, "/api/platform/pricing/gpt-4o/history", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("price history: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var history []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&history); err != nil {
		t.Fatal(err)
	}
	if len(history) < 2 {
		t.Fatalf("expected at least 2 history entries, got %d", len(history))
	}
	// History is DESC by effective_from — most recent first
	if history[0]["inputPrice"].(float64) != 1.5 {
		t.Errorf("latest inputPrice = %v, want 1.5", history[0]["inputPrice"])
	}
}

func TestPlatformContractPricing(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	cookie := saas.LoginPlatform(t, router)

	// Create a company
	created := saas.CreateCompanyHTTP(t, router, cookie, "Contract Co", "admin@contract.example")
	companyID := created.Company.ID.String()

	// Set contract pricing
	body, _ := json.Marshal(map[string]any{
		"modelType":   "deepseek-chat",
		"inputPrice":  0.8,
		"outputPrice": 3.2,
		"note":        "Q3 discount",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/platform/companies/"+companyID+"/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("set contract pricing: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// List contract pricing
	req = httptest.NewRequest(http.MethodGet, "/api/platform/companies/"+companyID+"/pricing", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list contract pricing: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var prices []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&prices); err != nil {
		t.Fatal(err)
	}
	if len(prices) != 1 {
		t.Fatalf("expected 1 contract price, got %d", len(prices))
	}
	if prices[0]["inputPrice"].(float64) != 0.8 {
		t.Errorf("contract inputPrice = %v, want 0.8", prices[0]["inputPrice"])
	}
}

func TestPlatformPricingUnauthorized(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/pricing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPlatformSetPricingValidation(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	cookie := saas.LoginPlatform(t, router)

	// Missing modelType
	body, _ := json.Marshal(map[string]any{
		"inputPrice":  1.0,
		"outputPrice": 4.0,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing modelType, got %d", rec.Code)
	}

	// Negative price
	body, _ = json.Marshal(map[string]any{
		"modelType":   "test-model",
		"inputPrice":  -1.0,
		"outputPrice": 4.0,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative price, got %d", rec.Code)
	}
}
