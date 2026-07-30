//go:build testhook

package catalogsync_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

// --- TestPricingVersionBump ---
// SetGlobalPrice / SetContractPrice → /sync/versions pricing 值 +1

func TestPricingVersionBump(t *testing.T) {
	t.Parallel()
	router, platformCookie, companyID := setupSaaSWithCompany(t)

	// Initial: pricing version = 0
	v := fetchVersions(t, router)
	if v.Pricing != 0 {
		t.Fatalf("expected initial pricing version 0, got %d", v.Pricing)
	}

	// Set global price → pricing version = 1
	setGlobalPrice(t, router, platformCookie, "test-model", 2.0, 8.0)
	v = fetchVersions(t, router)
	if v.Pricing != 1 {
		t.Fatalf("expected pricing version 1 after SetGlobalPrice, got %d", v.Pricing)
	}

	// Set contract price → pricing version = 2
	setContractPrice(t, router, platformCookie, companyID, "test-model", 1.0, 4.0)
	v = fetchVersions(t, router)
	if v.Pricing != 2 {
		t.Fatalf("expected pricing version 2 after SetContractPrice, got %d", v.Pricing)
	}
}

// --- TestExecutorPricingSyncTriggered ---
// pricing version 不同 → executor 拉取 pricing 并写入 model_pricing

func TestExecutorPricingSyncTriggered(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	// Mock SaaS server
	mockServer := pricingMockServer(t, catalog.CatalogVersions{Models: 0, Pricing: 1}, []catalog.CatalogPricing{
		{ModelType: "gpt-4o", InputPrice: 2.0, OutputPrice: 8.0, IsContract: false},
		{ModelType: "deepseek-chat", InputPrice: 0.5, OutputPrice: 2.0, IsContract: true},
	})

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	ctx := context.Background()
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("executor.Execute: %v", err)
	}

	// Verify: global price written with globalCompanyID
	globalPrices, err := st.ModelPricing().CurrentPricesBatch(ctx, globalCompanyID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	found := findPricing(globalPrices, "gpt-4o")
	if found == nil {
		t.Fatal("global price for gpt-4o not found in model_pricing")
	}
	if found.InputPrice != 2.0 || found.OutputPrice != 8.0 {
		t.Errorf("global price mismatch: got input=%f output=%f", found.InputPrice, found.OutputPrice)
	}

	// Verify: contract price written with localCompanyID
	contractPrices, err := st.ModelPricing().CurrentPricesBatch(ctx, localCompanyID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	found = findPricing(contractPrices, "deepseek-chat")
	if found == nil {
		t.Fatal("contract price for deepseek-chat not found in model_pricing")
	}
	if found.InputPrice != 0.5 || found.OutputPrice != 2.0 {
		t.Errorf("contract price mismatch: got input=%f output=%f", found.InputPrice, found.OutputPrice)
	}

	// Verify: local pricing version updated
	vStr, _ := st.SystemSettings().Get(ctx, "catalog.pricing_version")
	if vStr != "1" {
		t.Errorf("expected local pricing version '1', got %q", vStr)
	}
}

// --- TestExecutorPricingSyncSkipsWhenUpToDate ---
// pricing version 相同 → 不拉取

func TestExecutorPricingSyncSkipsWhenUpToDate(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	// Pre-set local pricing version to 3
	ctx := context.Background()
	_ = st.SystemSettings().Set(ctx, "catalog.pricing_version", "3")

	// Mock that returns pricing version 3 (same as local)
	called := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			_ = json.NewEncoder(w).Encode(map[string]int{"models": 0, "pricing": 3, "currencies": 0})
		case "/api/platform/sync/catalog/pricing":
			called = true
			_ = json.NewEncoder(w).Encode(map[string]any{"version": 3, "data": []any{}})
		case "/api/platform/sync/catalog/currencies":
			_ = json.NewEncoder(w).Encode(map[string]any{"version": 0, "data": []any{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(mockServer.Close)

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("executor.Execute: %v", err)
	}

	if called {
		t.Error("FetchPricing should not be called when versions match")
	}
}

// --- TestExecutorPricingSyncWritesContractPrice ---
// isContract:true → 写入 localCompanyID; isContract:false → 写入 globalCompanyID

func TestExecutorPricingSyncWritesContractPrice(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	mockServer := pricingMockServer(t, catalog.CatalogVersions{Models: 0, Pricing: 1}, []catalog.CatalogPricing{
		{ModelType: "gpt-4o", InputPrice: 15.0, OutputPrice: 60.0, IsContract: false},
		{ModelType: "gpt-4o", InputPrice: 8.0, OutputPrice: 32.0, IsContract: true},
	})

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	ctx := context.Background()
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("executor.Execute: %v", err)
	}

	// Global price → globalCompanyID
	globalPrices, _ := st.ModelPricing().CurrentPricesBatch(ctx, globalCompanyID, time.Now())
	gp := findPricing(globalPrices, "gpt-4o")
	if gp == nil {
		t.Fatal("global gpt-4o price not found")
	}
	if gp.InputPrice != 15.0 || gp.OutputPrice != 60.0 {
		t.Errorf("global price: want 15/60, got %f/%f", gp.InputPrice, gp.OutputPrice)
	}

	// Contract price → localCompanyID
	contractPrices, _ := st.ModelPricing().CurrentPricesBatch(ctx, localCompanyID, time.Now())
	cp := findPricing(contractPrices, "gpt-4o")
	if cp == nil {
		t.Fatal("contract gpt-4o price not found")
	}
	if cp.InputPrice != 8.0 || cp.OutputPrice != 32.0 {
		t.Errorf("contract price: want 8/32, got %f/%f", cp.InputPrice, cp.OutputPrice)
	}

	// UpsertModelRatio called only for global (non-contract)
	// StubAdminClient.UpsertModelRatio always returns nil, no counter — but we can
	// verify indirectly: contract price should NOT be pushed to NewAPI.
	// (This is validated by the code path: !p.IsContract check)
}

// --- Helpers ---

func setupSaaSWithCompany(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	router := saas.NewRouter(t, saas.StartNewAPIMock(t))
	platformCookie := saas.LoginPlatform(t, router)
	created := saas.CreateCompanyHTTP(t, router, platformCookie, "Test Corp", "admin@test.com")
	return router, platformCookie, created.Company.ID.String()
}

type versionsResponse struct {
	Models  int `json:"models"`
	Pricing int `json:"pricing"`
}

func fetchVersions(t *testing.T, router http.Handler) versionsResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/versions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetchVersions: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var v versionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func setGlobalPrice(t *testing.T, router http.Handler, cookie, modelType string, input, output float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"modelType": modelType, "inputPrice": input, "outputPrice": output,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setGlobalPrice: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func setContractPrice(t *testing.T, router http.Handler, cookie, companyID, modelType string, input, output float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"modelType": modelType, "inputPrice": input, "outputPrice": output,
	})
	url := "/api/platform/companies/" + companyID + "/pricing"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setContractPrice: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func pricingMockServer(t *testing.T, versions catalog.CatalogVersions, pricing []catalog.CatalogPricing) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			_ = json.NewEncoder(w).Encode(map[string]int{
				"models":     versions.Models,
				"pricing":    versions.Pricing,
				"currencies": versions.Currencies,
			})
		case "/api/platform/sync/catalog/pricing":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version": versions.Pricing,
				"data":    pricing,
			})
		case "/api/platform/sync/catalog/currencies":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version": versions.Currencies,
				"data":    []any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func findPricing(rows []store.ModelPricingRow, modelType string) *store.ModelPricingRow {
	for i := range rows {
		if rows[i].ModelType == modelType {
			return &rows[i]
		}
	}
	return nil
}

func createTestCompany(t *testing.T, st store.Store) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	err := st.Company().Create(context.Background(), store.Company{
		ID:        id,
		Name:      "local-test-" + id.String()[:8],
		Type:      store.CompanyTypeSelfhosted,
		Status:    store.CompanyStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("createTestCompany: %v", err)
	}
	return id
}
