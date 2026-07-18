package billing_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	saas "github.com/tokenjoy/backend/tests/testutil/saas"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestBillingWalletAfterPlatformRecharge(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Billing Co", "billing@example.com", "Billing Admin", "securepass123")

	saas.PlatformRechargeHTTP(t, router, platformCookie, provisioned.Company.ID, 100)

	req := httptest.NewRequest(http.MethodGet, "/api/billing/wallet", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var wallet domainbilling.WalletView
	if err := json.NewDecoder(rec.Body).Decode(&wallet); err != nil {
		t.Fatal(err)
	}
	if domainbilling.PrimaryWalletBalance(wallet) <= 0 {
		t.Fatalf("expected positive wallet balance after recharge, got %v", domainbilling.PrimaryWalletBalance(wallet))
	}
	if wallet.CompanyID != provisioned.Company.ID {
		t.Fatalf("expected companyId %d, got %d", provisioned.Company.ID, wallet.CompanyID)
	}
}

func TestBillingSelfRechargeConfirmFlow(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	router := app.Router
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Self Rch", "self-rch@example.com", "Self Admin", "securepass123")

	body, _ := json.Marshal(map[string]any{
		"amount": 50.0, "idempotencyKey": "self-recharge-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/recharge", bytes.NewReader(body))
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rec.Code, rec.Body.String())
	}
	var order store.RechargeOrder
	if err := json.NewDecoder(rec.Body).Decode(&order); err != nil {
		t.Fatal(err)
	}
	if order.Status != store.RechargeStatusPending {
		t.Fatalf("expected pending, got %s", order.Status)
	}

	confirmURL := fmt.Sprintf("/api/billing/recharge/%s/confirm", order.ID)
	req = httptest.NewRequest(http.MethodPost, confirmURL, nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("confirm: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}

	stored, err := app.Store.Billing().GetRechargeOrder(testutil.CtxForCompany(provisioned.Company.ID), order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != store.RechargeStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", stored.Status)
	}
}

func TestBillingSelfRechargeIdempotencyKey(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Idem Co", "idem@example.com", "Idem Admin", "securepass123")

	payload, _ := json.Marshal(map[string]any{"amount": 10.0, "idempotencyKey": "unique-key-42"})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/recharge", bytes.NewReader(payload))
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("first create: expected 202, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/billing/recharge", bytes.NewReader(payload))
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusAccepted {
		t.Fatal("expected error for duplicate idempotency key")
	}
}

func TestBillingWalletUsesNewAPIQuota(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Quota Co", "quota@example.com", "Quota Admin", "securepass123")

	saas.PlatformRechargeHTTP(t, router, platformCookie, provisioned.Company.ID, 200)

	req := httptest.NewRequest(http.MethodGet, "/api/billing/wallet", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	var wallet domainbilling.WalletView
	if err := json.NewDecoder(rec.Body).Decode(&wallet); err != nil {
		t.Fatal(err)
	}
	balance := domainbilling.PrimaryWalletBalance(wallet)
	if balance < 199 || balance > 201 {
		t.Fatalf("expected balance ~200 after recharge, got %v", balance)
	}
}
