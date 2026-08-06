//go:build testhook

package catalogsync_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

// TestCurrenciesSyncTriggered — version 不同 → 拉取 + ReplaceCurrencies
func TestCurrenciesSyncTriggered(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	mockServer := currenciesMockServer(t, catalog.CatalogVersions{Models: 1, Pricing: 1, Currencies: 2}, []catalog.CatalogCurrency{
		{Code: "CNY", QuotaPerUnit: 500000},
		{Code: "USD", QuotaPerUnit: 3600000},
	})

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	ctx := context.Background()
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("executor.Execute: %v", err)
	}

	// Verify: currencies table has CNY and USD, both enabled
	cny, err := st.Billing().GetCurrency(ctx, "CNY")
	if err != nil {
		t.Fatal(err)
	}
	if cny == nil {
		t.Fatal("CNY not found in currencies")
	}
	if cny.QuotaPerUnit != 500000 || !cny.Enabled {
		t.Errorf("CNY: got qpu=%d enabled=%v, want 500000/true", cny.QuotaPerUnit, cny.Enabled)
	}

	usd, err := st.Billing().GetCurrency(ctx, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if usd == nil {
		t.Fatal("USD not found in currencies")
	}
	if usd.QuotaPerUnit != 3600000 || !usd.Enabled {
		t.Errorf("USD: got qpu=%d enabled=%v, want 3600000/true", usd.QuotaPerUnit, usd.Enabled)
	}

	// Verify: sync_versions currencies updated
	cv, _ := st.SyncVersions().Get(ctx, store.GlobalSyncVersion, "currencies")
	if cv != 2 {
		t.Errorf("expected local currencies version 2, got %d", cv)
	}
}

// TestCurrenciesSyncDisablesStale — 远端移除币种 → 本地 disable
func TestCurrenciesSyncDisablesStale(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	ctx := context.Background()

	// Pre-insert CNY + USD as enabled
	_ = st.Billing().UpsertCurrency(ctx, store.Currency{Code: "CNY", QuotaPerUnit: 500000, Enabled: true})
	_ = st.Billing().UpsertCurrency(ctx, store.Currency{Code: "USD", QuotaPerUnit: 3600000, Enabled: true})

	// Mock returns only CNY — USD should get disabled
	mockServer := currenciesMockServer(t, catalog.CatalogVersions{Models: 1, Pricing: 1, Currencies: 2}, []catalog.CatalogCurrency{
		{Code: "CNY", QuotaPerUnit: 500000},
	})

	stub := &mock.StubAdminClient{}
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL})
	executor := catalogsync.NewExecutor(client, stub, st, globalCompanyID, localCompanyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("executor.Execute: %v", err)
	}

	// Verify: CNY still enabled, USD disabled
	cny, _ := st.Billing().GetCurrency(ctx, "CNY")
	if cny == nil || !cny.Enabled {
		t.Error("CNY should still be enabled")
	}

	usd, _ := st.Billing().GetCurrency(ctx, "USD")
	if usd == nil {
		t.Fatal("USD should still exist (not deleted)")
	}
	if usd.Enabled {
		t.Error("USD should be disabled after sync")
	}
}

// TestCurrenciesSyncSkipsWhenUpToDate — version 相同 → 不拉取
func TestCurrenciesSyncSkipsWhenUpToDate(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)

	globalCompanyID := cfg.TokenJoyCompanyID
	localCompanyID := createTestCompany(t, st)

	ctx := context.Background()
	// Pre-set local currencies version to 2
	_ = st.SyncVersions().Set(ctx, store.GlobalSyncVersion, "currencies", 2)

	called := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			_ = json.NewEncoder(w).Encode(map[string]int{"models": 1, "pricing": 1, "currencies": 2})
		case "/api/platform/sync/catalog/currencies":
			called = true
			_ = json.NewEncoder(w).Encode(map[string]any{"version": 2, "data": []any{}})
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
		t.Error("FetchCurrencies should not be called when versions match")
	}
}

// --- Helpers ---

func currenciesMockServer(t *testing.T, versions catalog.CatalogVersions, currencies []catalog.CatalogCurrency) *httptest.Server {
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
		case "/api/platform/sync/catalog/currencies":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version": versions.Currencies,
				"data":    currencies,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
