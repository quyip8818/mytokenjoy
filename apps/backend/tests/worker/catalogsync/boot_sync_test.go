//go:build testhook

package catalogsync_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/support/tenant"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

// TestBootSyncAllChannels verifies that Execute on a fresh DB (all versions=0)
// syncs models, pricing, and currencies in a single call — the "boot-time sync" scenario.
func TestBootSyncAllChannels(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	models := []catalog.CatalogModel{
		{ModelID: "gpt-4o", DisplayName: "GPT-4o", Provider: "openai", Capabilities: []string{"chat"}, MaxContext: 128000},
		{ModelID: "claude-3.5-sonnet", DisplayName: "Claude 3.5 Sonnet", Provider: "anthropic", Capabilities: []string{"chat"}, MaxContext: 200000},
	}
	pricing := []catalog.CatalogPricing{
		{ModelType: "gpt-4o", InputPrice: 2.5, OutputPrice: 10.0},
	}
	currencies := []catalog.CatalogCurrency{
		{Code: "CNY", QuotaPerUnit: 500000},
	}

	mockServer := fullCatalogMockServer(t, catalog.CatalogVersions{Models: 2, Pricing: 2, Currencies: 2}, models, pricing, currencies)

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	ctx := context.Background()
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// --- Models: verify synced to globalCompanyID ---
	modelCtx := tenant.With(ctx, tenant.Info{CompanyID: globalCompanyID})
	allModels, err := st.Models().Models(modelCtx)
	if err != nil {
		t.Fatal(err)
	}
	if len(allModels) < 2 {
		t.Fatalf("expected at least 2 models, got %d", len(allModels))
	}
	var foundGPT, foundClaude bool
	for _, m := range allModels {
		if m.Type == "gpt-4o" && m.Provider == "openai" && !m.Deprecated {
			foundGPT = true
		}
		if m.Type == "claude-3.5-sonnet" && m.Provider == "anthropic" && !m.Deprecated {
			foundClaude = true
		}
	}
	if !foundGPT {
		t.Error("gpt-4o model not found after boot sync")
	}
	if !foundClaude {
		t.Error("claude-3.5-sonnet model not found after boot sync")
	}

	// --- Pricing: verify written ---
	vStr, _ := st.SystemSettings().Get(ctx, "catalog.pricing_version")
	if vStr != "2" {
		t.Errorf("pricing version: want '2', got %q", vStr)
	}

	// --- Currencies: verify written ---
	vStr, _ = st.SystemSettings().Get(ctx, "catalog.currencies_version")
	if vStr != "2" {
		t.Errorf("currencies version: want '2', got %q", vStr)
	}
	cny, err := st.Billing().GetCurrency(ctx, "CNY")
	if err != nil || cny == nil {
		t.Fatal("CNY currency not found after boot sync")
	}

	// --- Models version: verify stored ---
	vStr, _ = st.SystemSettings().Get(ctx, "catalog.models_version")
	if vStr != "2" {
		t.Errorf("models version: want '2', got %q", vStr)
	}
}

// TestBootSyncSkipsWhenAlreadySynced — all versions match → no fetch calls made.
func TestBootSyncSkipsWhenAlreadySynced(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	ctx := context.Background()
	_ = st.SystemSettings().Set(ctx, "catalog.models_version", "5")
	_ = st.SystemSettings().Set(ctx, "catalog.pricing_version", "5")
	_ = st.SystemSettings().Set(ctx, "catalog.currencies_version", "5")

	var fetchCalls int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			_ = json.NewEncoder(w).Encode(map[string]int{"models": 5, "pricing": 5, "currencies": 5})
		default:
			fetchCalls++
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(mockServer.Close)

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if fetchCalls > 0 {
		t.Errorf("expected 0 fetch calls when versions match, got %d", fetchCalls)
	}
}

// --- Helper ---

func fullCatalogMockServer(t *testing.T, versions catalog.CatalogVersions, models []catalog.CatalogModel, pricing []catalog.CatalogPricing, currencies []catalog.CatalogCurrency) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			_ = json.NewEncoder(w).Encode(map[string]int{
				"models":     versions.Models,
				"pricing":    versions.Pricing,
				"currencies": versions.Currencies,
				"walletLots": versions.WalletLots,
			})
		case "/api/platform/sync/catalog/models":
			_ = json.NewEncoder(w).Encode(map[string]any{"version": versions.Models, "data": models})
		case "/api/platform/sync/catalog/pricing":
			_ = json.NewEncoder(w).Encode(map[string]any{"version": versions.Pricing, "data": pricing})
		case "/api/platform/sync/catalog/currencies":
			_ = json.NewEncoder(w).Encode(map[string]any{"version": versions.Currencies, "data": currencies})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
