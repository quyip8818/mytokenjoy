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

// --- CatalogLots endpoint tests ---

func TestCatalogLotsRequiresSyncToken(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/wallet_lots", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403 without token, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCatalogLotsWithValidToken(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	syncToken := registerAndGetToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/wallet_lots", nil)
	req.Header.Set("Authorization", "Bearer "+syncToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data              []json.RawMessage `json:"data"`
		WalletRemainQuota int64             `json:"walletRemainQuota"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	// Newly created company has no lots yet → empty list, wallet=0.
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 lots for new company, got %d", len(resp.Data))
	}
	if resp.WalletRemainQuota != 0 {
		t.Fatalf("expected walletRemainQuota=0 for new company, got %d", resp.WalletRemainQuota)
	}
}

func TestCatalogLotsReturnsLotsAfterRecharge(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	// Register and get both sync token and company ID.
	result := registerAndGetAll(t, router)

	// Recharge the company via platform admin.
	platformCookie := saas.LoginPlatform(t, router)
	body, _ := json.Marshal(map[string]float64{"amount": 100.0})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/companies/"+result.CompanyID+"/recharge", bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("recharge: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Fetch lots with sync token.
	req = httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/wallet_lots", nil)
	req.Header.Set("Authorization", "Bearer "+result.SyncToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var lotsResp struct {
		Data              []map[string]any `json:"data"`
		Orders            []map[string]any `json:"orders"`
		WalletRemainQuota int64            `json:"walletRemainQuota"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&lotsResp); err != nil {
		t.Fatal(err)
	}

	if len(lotsResp.Data) == 0 {
		t.Fatal("expected at least 1 lot after recharge")
	}
	if lotsResp.WalletRemainQuota <= 0 {
		t.Fatalf("expected positive walletRemainQuota after recharge, got %d", lotsResp.WalletRemainQuota)
	}

	// Verify lot has orderId field.
	lot := lotsResp.Data[0]
	for _, field := range []string{"id", "orderId", "lotKind", "billingCurrency", "quotaPerUnit", "quotaGranted", "quotaRemaining", "status", "createdAt"} {
		if _, ok := lot[field]; !ok {
			t.Fatalf("lot missing field %q", field)
		}
	}
	if lot["orderId"] == "" {
		t.Fatal("lot orderId should not be empty")
	}

	// Verify orders array is populated.
	if len(lotsResp.Orders) == 0 {
		t.Fatal("expected at least 1 order after recharge")
	}
	order := lotsResp.Orders[0]
	for _, field := range []string{"id", "amount", "currency", "source", "lotKind", "status", "createdAt"} {
		if _, ok := order[field]; !ok {
			t.Fatalf("order missing field %q", field)
		}
	}
	// Verify lot.orderId matches an order in the orders array.
	lotOrderID := lot["orderId"].(string)
	foundOrder := false
	for _, o := range lotsResp.Orders {
		if o["id"] == lotOrderID {
			foundOrder = true
			break
		}
	}
	if !foundOrder {
		t.Fatalf("lot orderId %q not found in orders array", lotOrderID)
	}
}
