//go:build testhook

package platform_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

// TestCurrencyCRUD — 完整生命周期
func TestCurrencyCRUD(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)
	platformCookie := saas.LoginPlatform(t, router)

	// 1. GET /platform/currencies → initial seed has CNY
	currencies := listCurrencies(t, router, platformCookie)
	if len(currencies) == 0 {
		t.Fatal("expected at least CNY from seed")
	}
	foundCNY := false
	for _, c := range currencies {
		if c.Code == "CNY" {
			foundCNY = true
		}
	}
	if !foundCNY {
		t.Fatal("CNY not found in initial list")
	}

	// 2. POST /platform/currencies — create USD
	body, _ := json.Marshal(map[string]any{"code": "USD", "quotaPerUnit": 3600000})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/currencies", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create USD: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// 3. GET → includes USD
	currencies = listCurrencies(t, router, platformCookie)
	foundUSD := false
	for _, c := range currencies {
		if c.Code == "USD" && c.QuotaPerUnit == 3600000 && c.Enabled {
			foundUSD = true
		}
	}
	if !foundUSD {
		t.Fatal("USD not found after creation")
	}

	// 4. PUT /platform/currencies/USD — update QPU
	body, _ = json.Marshal(map[string]any{"quotaPerUnit": 3700000})
	req = httptest.NewRequest(http.MethodPut, "/api/platform/currencies/USD", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update USD: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// 5. GET → USD.quotaPerUnit = 3700000
	currencies = listCurrencies(t, router, platformCookie)
	for _, c := range currencies {
		if c.Code == "USD" {
			if c.QuotaPerUnit != 3700000 {
				t.Errorf("USD QPU: got %d, want 3700000", c.QuotaPerUnit)
			}
		}
	}

	// 6. Verify: currencies version incremented (3 writes: create + update + seed already has version 1)
	v := fetchCurrenciesVersion(t, router)
	if v < 2 {
		t.Errorf("expected currencies version >= 2 after create+update, got %d", v)
	}
}

// TestDisableCurrencyBlockedByFK — 被企业引用的币种禁用返回 409
func TestDisableCurrencyBlockedByFK(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)

	// Create a company (uses CNY as billing_currency by default)
	saas.CreateCompanyHTTP(t, router, platformCookie, "FK Test Co", "admin@fk.example")

	// Try to disable CNY — should fail with 409
	body, _ := json.Marshal(map[string]any{"enabled": false})
	req := httptest.NewRequest(http.MethodPatch, "/api/platform/currencies/CNY/status", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("disable CNY: expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify CNY still enabled
	currencies := listCurrencies(t, router, platformCookie)
	for _, c := range currencies {
		if c.Code == "CNY" && !c.Enabled {
			t.Error("CNY should still be enabled after failed disable")
		}
	}
}

// TestCurrencyUpdatedByName — 写操作后 updatedByName 返回操作人名字
func TestCurrencyUpdatedByName(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)
	platformCookie := saas.LoginPlatform(t, router)

	// Seed CNY should have updatedByName = nil (no actor on bootstrap)
	currencies := listCurrencies(t, router, platformCookie)
	for _, c := range currencies {
		if c.Code == "CNY" {
			if c.UpdatedByName != nil {
				t.Errorf("CNY from seed should have updatedByName=nil, got %q", *c.UpdatedByName)
			}
			break
		}
	}

	// Create JPY — should have updatedByName set
	body, _ := json.Marshal(map[string]any{"code": "JPY", "quotaPerUnit": 7000})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/currencies", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create JPY: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// List — JPY should have updatedByName set
	currencies = listCurrencies(t, router, platformCookie)
	for _, c := range currencies {
		if c.Code == "JPY" {
			if c.UpdatedByName == nil || *c.UpdatedByName == "" {
				t.Error("JPY.updatedByName should be set after creation by platform admin")
			}
			return
		}
	}
	t.Fatal("JPY not found after creation")
}

// TestCatalogCurrenciesEndpoint — sync endpoint 返回正确格式
func TestCatalogCurrenciesEndpoint(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)

	// GET /sync/catalog/currencies — no auth required
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/currencies", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("catalog currencies: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Version int `json:"version"`
		Data    []struct {
			Code         string `json:"code"`
			QuotaPerUnit int64  `json:"quotaPerUnit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Should contain seed CNY
	if len(resp.Data) == 0 {
		t.Fatal("expected at least one currency in sync response")
	}
	foundCNY := false
	for _, c := range resp.Data {
		if c.Code == "CNY" && c.QuotaPerUnit > 0 {
			foundCNY = true
		}
	}
	if !foundCNY {
		t.Error("CNY not found in sync catalog response")
	}

	// Version should be >= 1 (seeded by insertSeedCurrencies)
	if resp.Version < 1 {
		t.Errorf("expected version >= 1, got %d", resp.Version)
	}
}

// --- Helpers ---

type currencyDTO struct {
	Code          string  `json:"code"`
	QuotaPerUnit  int64   `json:"quotaPerUnit"`
	Enabled       bool    `json:"enabled"`
	UpdatedAt     string  `json:"updatedAt"`
	UpdatedByName *string `json:"updatedByName"`
}

func listCurrencies(t *testing.T, router http.Handler, cookie string) []currencyDTO {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/platform/currencies", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listCurrencies: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []currencyDTO
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode currencies: %v", err)
	}
	return out
}

func fetchCurrenciesVersion(t *testing.T, router http.Handler) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/currencies", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetchCurrenciesVersion: expected 200, got %d", rec.Code)
	}
	var v struct {
		Version int `json:"version"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&v)
	return v.Version
}
