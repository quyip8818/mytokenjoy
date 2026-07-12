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

func TestWalletIncludesUsageStats(t *testing.T) {
	t.Parallel()
	app := testhttp.NewApp(t, nil)
	testutil.ApplyDemoRuntime(t, app.Store, app.Config)
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
	primary := domainbilling.PrimaryWalletBalance(wallet)
	if primary <= 0 {
		t.Fatalf("expected positive wallet balance, got %v", primary)
	}
	if wallet.TotalRequests <= 0 {
		t.Fatalf("expected positive totalRequests, got %d", wallet.TotalRequests)
	}
}
