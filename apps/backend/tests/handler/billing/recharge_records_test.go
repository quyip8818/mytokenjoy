package billing_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestListRechargeRecordsHTTP(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/billing/recharge-records", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var records []domainbilling.RechargeRecord
	if err := json.NewDecoder(rec.Body).Decode(&records); err != nil {
		t.Fatal(err)
	}
	if len(records) < 5 {
		t.Fatalf("expected at least 5 seeded records, got %d", len(records))
	}
}

func TestWalletIncludesUsageStats(t *testing.T) {
	app := testutil.NewTestApp(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/billing/wallet", nil)
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var wallet domainbilling.WalletView
	if err := json.NewDecoder(rec.Body).Decode(&wallet); err != nil {
		t.Fatal(err)
	}
	if wallet.TotalConsumed <= 0 {
		t.Fatalf("expected positive totalConsumed, got %v", wallet.TotalConsumed)
	}
	if wallet.TotalRequests <= 0 {
		t.Fatalf("expected positive totalRequests, got %d", wallet.TotalRequests)
	}
}
