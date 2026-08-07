//go:build testhook

package platform_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestPlatformListCompanyLots(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	created := saas.CreateCompanyHTTP(t, router, platformCookie, "Lot Audit Co", "admin@lotaudit.example")

	saas.PlatformRechargeHTTP(t, router, platformCookie, created.Company.ID, 100)

	// List lots.
	url := fmt.Sprintf("/api/platform/companies/%s/lots", created.Company.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Lots              []domainbilling.LotAuditEntry `json:"lots"`
		WalletRemainQuota int64                         `json:"walletRemainQuota"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Lots) == 0 {
		t.Fatal("expected at least 1 lot after recharge")
	}
	if resp.WalletRemainQuota <= 0 {
		t.Fatalf("expected positive walletRemainQuota, got %d", resp.WalletRemainQuota)
	}

	// Verify lot has a credit transaction.
	lot := resp.Lots[0]
	if lot.Status != "active" {
		t.Fatalf("expected active lot, got %s", lot.Status)
	}
	if len(lot.Transactions) == 0 {
		t.Fatal("expected credit transaction on lot after recharge")
	}
	tx := lot.Transactions[0]
	if tx.Action != "credit" {
		t.Fatalf("expected credit action, got %s", tx.Action)
	}
	if tx.QuotaDelta <= 0 {
		t.Fatalf("expected positive quota delta, got %d", tx.QuotaDelta)
	}
}

func TestPlatformRefundCompany(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	created := saas.CreateCompanyHTTP(t, router, platformCookie, "Refund Co", "admin@refund.example")

	saas.PlatformRechargeHTTP(t, router, platformCookie, created.Company.ID, 100)

	// Get lots to find the lot ID.
	lotsURL := fmt.Sprintf("/api/platform/companies/%s/lots", created.Company.ID)
	req := httptest.NewRequest(http.MethodGet, lotsURL, nil)
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list lots: expected 200, got %d", rec.Code)
	}
	var lotsResp struct {
		Lots              []domainbilling.LotAuditEntry `json:"lots"`
		WalletRemainQuota int64                         `json:"walletRemainQuota"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&lotsResp); err != nil {
		t.Fatal(err)
	}
	if len(lotsResp.Lots) == 0 {
		t.Fatal("no lots found")
	}
	lotID := lotsResp.Lots[0].ID.String()
	walletBefore := lotsResp.WalletRemainQuota

	// Refund ¥30 from the lot.
	refundURL := fmt.Sprintf("/api/platform/companies/%s/refund", created.Company.ID)
	body, _ := json.Marshal(map[string]any{"lotId": lotID, "amount": 30.0})
	req = httptest.NewRequest(http.MethodPost, refundURL, bytes.NewReader(body))
	req.Header.Set("Cookie", platformCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("refund: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify lot remaining decreased and refund transaction exists.
	req = httptest.NewRequest(http.MethodGet, lotsURL, nil)
	req.Header.Set("Cookie", platformCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list lots after refund: expected 200, got %d", rec.Code)
	}
	var afterResp struct {
		Lots              []domainbilling.LotAuditEntry `json:"lots"`
		WalletRemainQuota int64                         `json:"walletRemainQuota"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&afterResp); err != nil {
		t.Fatal(err)
	}
	if afterResp.WalletRemainQuota >= walletBefore {
		t.Fatalf("expected wallet to decrease after refund: before=%d after=%d", walletBefore, afterResp.WalletRemainQuota)
	}

	// Find the lot and check transactions.
	var found *domainbilling.LotAuditEntry
	for i := range afterResp.Lots {
		if afterResp.Lots[i].ID.String() == lotID {
			found = &afterResp.Lots[i]
			break
		}
	}
	if found == nil {
		t.Fatal("lot not found after refund")
	}
	if found.QuotaRemaining >= lotsResp.Lots[0].QuotaRemaining {
		t.Fatalf("expected lot remaining to decrease: before=%d after=%d", lotsResp.Lots[0].QuotaRemaining, found.QuotaRemaining)
	}

	// Should have 2 transactions: credit + refund.
	if len(found.Transactions) < 2 {
		t.Fatalf("expected at least 2 transactions (credit + refund), got %d", len(found.Transactions))
	}
	var hasRefund bool
	for _, tx := range found.Transactions {
		if tx.Action == "refund" {
			hasRefund = true
			if tx.QuotaDelta >= 0 {
				t.Fatalf("refund transaction should have negative delta, got %d", tx.QuotaDelta)
			}
		}
	}
	if !hasRefund {
		t.Fatal("expected a refund transaction")
	}
}
