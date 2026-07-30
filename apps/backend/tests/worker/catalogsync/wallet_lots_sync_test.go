//go:build testhook

package catalogsync_test

import (
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
)

// --- wallet_lots sync: lot mirroring ---

// TestWalletLotsSyncMirrorsLotsFromSaaS verifies that syncWalletLots
// upserts lot + order rows from the SaaS response into the local DB.
func TestWalletLotsSyncMirrorsLotsFromSaaS(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t, testutil.WithIngestEnabled(true))
	ctx := context.Background()

	companyID := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	_ = st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Lot Sync Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	orderID := uuid.Must(uuid.NewV7())
	lotID := uuid.Must(uuid.NewV7())

	mockServer := walletLotsMockServer(t, 2, []catalog.CatalogOrder{
		{
			ID: orderID.String(), Amount: 100.0, Currency: "CNY",
			QuotaPerUnit: 500000, QuotaGranted: 50000000,
			Source: "self", LotKind: "paid", Status: "confirmed",
			DisplayOrderID: "ORD-001", PaymentMethod: "alipay",
			CreatedAt: now.Unix(),
		},
	}, []catalog.CatalogLot{
		{
			ID: lotID.String(), OrderID: orderID.String(),
			LotKind: "paid", BillingCurrency: "CNY",
			QuotaPerUnit: 500000, QuotaGranted: 50000000, QuotaRemaining: 40000000,
			PaidAmount: 100.0, Status: "active", CreatedAt: now.Unix(),
		},
	}, 40000000)

	globalID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL, SyncToken: "test"})
	executor := catalogsync.NewExecutor(client, &mock.StubAdminClient{}, st, globalID, companyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify order was created.
	order, err := st.Billing().GetRechargeOrder(ctx, orderID)
	if err != nil || order == nil {
		t.Fatalf("order not synced: %v", err)
	}
	if order.Amount != 100.0 {
		t.Fatalf("order amount: got %f want 100.0", order.Amount)
	}
	if order.Source != "self" {
		t.Fatalf("order source: got %q want 'self'", order.Source)
	}

	// Verify lot was created.
	lot, err := st.Billing().GetLotByID(ctx, lotID)
	if err != nil || lot == nil {
		t.Fatalf("lot not synced: %v", err)
	}
	if lot.QuotaRemaining != 40000000 {
		t.Fatalf("lot quota_remaining: got %d want 40000000", lot.QuotaRemaining)
	}
	if lot.RechargeOrderID != orderID {
		t.Fatalf("lot order reference: got %s want %s", lot.RechargeOrderID, orderID)
	}
}

// TestWalletLotsSyncUpdatesExistingLot verifies that a second sync
// updates quota_remaining and status on an existing lot (not duplicate insert).
func TestWalletLotsSyncUpdatesExistingLot(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t, testutil.WithIngestEnabled(true))
	ctx := context.Background()

	companyID := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	_ = st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Update Sync Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	orderID := uuid.Must(uuid.NewV7())
	lotID := uuid.Must(uuid.NewV7())
	globalID := uuid.MustParse("00000000-0000-7000-8000-000000000001")

	// First sync: lot has 50M remaining.
	mock1 := walletLotsMockServer(t, 1, []catalog.CatalogOrder{
		{ID: orderID.String(), Amount: 100.0, Currency: "CNY", QuotaPerUnit: 500000, QuotaGranted: 50000000, Source: "self", LotKind: "paid", Status: "confirmed", CreatedAt: now.Unix()},
	}, []catalog.CatalogLot{
		{ID: lotID.String(), OrderID: orderID.String(), LotKind: "paid", BillingCurrency: "CNY", QuotaPerUnit: 500000, QuotaGranted: 50000000, QuotaRemaining: 50000000, PaidAmount: 100.0, Status: "active", CreatedAt: now.Unix()},
	}, 50000000)

	client1 := catalog.NewClient(catalog.Config{BaseURL: mock1.URL, SyncToken: "test"})
	exec1 := catalogsync.NewExecutor(client1, &mock.StubAdminClient{}, st, globalID, companyID)
	if err := exec1.Execute(ctx); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	lot, _ := st.Billing().GetLotByID(ctx, lotID)
	if lot.QuotaRemaining != 50000000 {
		t.Fatalf("after first sync: got %d want 50000000", lot.QuotaRemaining)
	}

	// Second sync: lot consumed down to 30M.
	mock2 := walletLotsMockServer(t, 2, []catalog.CatalogOrder{
		{ID: orderID.String(), Amount: 100.0, Currency: "CNY", QuotaPerUnit: 500000, QuotaGranted: 50000000, Source: "self", LotKind: "paid", Status: "confirmed", CreatedAt: now.Unix()},
	}, []catalog.CatalogLot{
		{ID: lotID.String(), OrderID: orderID.String(), LotKind: "paid", BillingCurrency: "CNY", QuotaPerUnit: 500000, QuotaGranted: 50000000, QuotaRemaining: 30000000, PaidAmount: 100.0, Status: "active", CreatedAt: now.Unix()},
	}, 30000000)

	client2 := catalog.NewClient(catalog.Config{BaseURL: mock2.URL, SyncToken: "test"})
	exec2 := catalogsync.NewExecutor(client2, &mock.StubAdminClient{}, st, globalID, companyID)
	if err := exec2.Execute(ctx); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	lot, _ = st.Billing().GetLotByID(ctx, lotID)
	if lot.QuotaRemaining != 30000000 {
		t.Fatalf("after second sync: got %d want 30000000", lot.QuotaRemaining)
	}
}

// --- wallet remain sync: wallet_remain_quota reconciliation ---

// TestWalletRemainSyncOverwritesLocalValue verifies that Catalog Sync
// overwrites the local wallet_remain_quota with the SaaS authoritative value.
func TestWalletRemainSyncOverwritesLocalValue(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t, testutil.WithIngestEnabled(true))
	ctx := context.Background()

	companyID := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	_ = st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Wallet Remain Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, WalletRemainQuota: 99999, // stale local value
		CreatedAt: now, UpdatedAt: now,
	})

	// SaaS says wallet = 12345678 (the authoritative value).
	saasWallet := int64(12345678)
	mockServer := walletLotsMockServer(t, 1, nil, nil, saasWallet)

	globalID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL, SyncToken: "test"})
	executor := catalogsync.NewExecutor(client, &mock.StubAdminClient{}, st, globalID, companyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify local wallet was overwritten with SaaS value.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal(err)
	}
	if co.WalletRemainQuota != saasWallet {
		t.Fatalf("wallet_remain_quota: got %d want %d", co.WalletRemainQuota, saasWallet)
	}
}

// TestWalletRemainSyncZeroIsValid verifies that wallet=0 from SaaS
// correctly sets local wallet to 0 (not skipped as "no data").
func TestWalletRemainSyncZeroIsValid(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t, testutil.WithIngestEnabled(true))
	ctx := context.Background()

	companyID := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	_ = st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Zero Wallet Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, WalletRemainQuota: 50000, // has balance
		CreatedAt: now, UpdatedAt: now,
	})

	// SaaS says wallet = 0 (company spent everything).
	mockServer := walletLotsMockServer(t, 1, nil, nil, 0)

	globalID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	client := catalog.NewClient(catalog.Config{BaseURL: mockServer.URL, SyncToken: "test"})
	executor := catalogsync.NewExecutor(client, &mock.StubAdminClient{}, st, globalID, companyID)

	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	co, _ := st.Company().GetByID(ctx, companyID)
	if co.WalletRemainQuota != 0 {
		t.Fatalf("wallet should be 0, got %d", co.WalletRemainQuota)
	}
}

// --- Helper ---

func walletLotsMockServer(t *testing.T, walletLotsVersion int, orders []catalog.CatalogOrder, lots []catalog.CatalogLot, walletRemain int64) *httptest.Server {
	t.Helper()
	if orders == nil {
		orders = []catalog.CatalogOrder{}
	}
	if lots == nil {
		lots = []catalog.CatalogLot{}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/platform/sync/versions":
			// Return versions matching seed (models=1, pricing=1, currencies=1) so those are skipped.
			_ = json.NewEncoder(w).Encode(map[string]int{
				"models": 1, "pricing": 1, "currencies": 1, "walletLots": walletLotsVersion,
			})
		case "/api/platform/sync/catalog/wallet_lots":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version":           walletLotsVersion,
				"data":              lots,
				"orders":            orders,
				"walletRemainQuota": walletRemain,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
