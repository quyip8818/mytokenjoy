package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOnboardingPlatformCreateAcceptInviteSession(t *testing.T) {
	mock := testutil.StartNewAPIMock(t)
	router := saasApp(t, mock)
	platformCookie := testutil.LoginPlatform(t, router)

	created := testutil.CreateCompanyHTTP(t, router, platformCookie,
		"onboard-co", "Onboard Co", "founder@onboard.example")
	member, memberCookie := testutil.AcceptInviteHTTP(t, router, created.InviteToken,
		"Founder", "securepass123")

	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.Header.Set("Cookie", memberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var session types.SessionContext
	if err := json.NewDecoder(rec.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if session.CompanyID != created.Company.ID {
		t.Fatalf("expected companyId %d, got %d", created.Company.ID, session.CompanyID)
	}
	if session.Member.ID != member.ID {
		t.Fatalf("expected member %s, got %s", member.ID, session.Member.ID)
	}
}

func TestOnboardingRejectSecondAcceptInvite(t *testing.T) {
	mock := testutil.StartNewAPIMock(t)
	router := saasApp(t, mock)
	platformCookie := testutil.LoginPlatform(t, router)

	created := testutil.CreateCompanyHTTP(t, router, platformCookie,
		"once-co", "Once Co", "once@example.com")
	_, _ = testutil.AcceptInviteHTTP(t, router, created.InviteToken, "Admin", "securepass123")

	body, _ := json.Marshal(map[string]string{
		"token": created.InviteToken, "name": "Other", "password": "securepass456",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for reused invite, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAcceptInviteHTTPRejectsShortPassword(t *testing.T) {
	mock := testutil.StartNewAPIMock(t)
	router := saasApp(t, mock)
	platformCookie := testutil.LoginPlatform(t, router)
	created := testutil.CreateCompanyHTTP(t, router, platformCookie,
		"short-pw-co", "Short PW", "short@example.com")

	body, _ := json.Marshal(map[string]string{
		"token": created.InviteToken, "name": "Admin", "password": "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestOnboardingWalletAndBudgetDualAxisGateway(t *testing.T) {
	mock := testutil.StartNewAPIMock(t)
	app := newTestApp(t, func(cfg *config.Config) {
		testutil.ApplySaaSConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.RelayGatewayEnabled = true
	})
	router := app.Router
	platformCookie := testutil.LoginPlatform(t, router)
	provisioned := testutil.ProvisionCompanyHTTP(t, router, platformCookie,
		"dual-axis", "Dual Axis", "dual@example.com", "Dual Admin", "securepass123")

	walletID := int64(0)
	if provisioned.Company.NewAPIWalletUserID != nil {
		walletID = *provisioned.Company.NewAPIWalletUserID
	}
	rootDept := fmt.Sprintf("dept-root-%d", provisioned.Company.ID)

	// No recharge: wallet 0 -> 403
	fullKey := testutil.ConfigureGatewayStore(t, app.Store, testutil.GatewayScenarioOpts{
		CompanyID:          provisioned.Company.ID,
		NewAPIWalletUserID: walletID,
		WalletQuota:        0,
		DepartmentID:       rootDept,
		Budget:             1000,
	})
	rec := httptest.NewRecorder()
	app.Router.ServeHTTP(rec, gatewayRequest(fullKey))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 with empty wallet, got %d", rec.Code)
	}

	// Recharge wallet but budget still 0 -> 403
	testutil.PlatformRechargeHTTP(t, router, platformCookie, provisioned.Company.ID, 100)
	mock.SetQuota(walletID, newapi.ToNewAPIUnits(100, nil, nil))
	fullKey = testutil.ConfigureGatewayStore(t, app.Store, testutil.GatewayScenarioOpts{
		CompanyID:          provisioned.Company.ID,
		NewAPIWalletUserID: walletID,
		WalletQuota:        newapi.ToNewAPIUnits(100, nil, nil),
		DepartmentID:       rootDept,
		Budget:             0,
		UseRealWallet:      false,
	})
	rec = httptest.NewRecorder()
	app.Router.ServeHTTP(rec, gatewayRequest(fullKey))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 with zero budget, got %d", rec.Code)
	}

	// Both wallet and budget -> 200
	testutil.UpdateBudgetNodeHTTP(t, router, provisioned.MemberCookie, rootDept, 1000)
	fullKey = testutil.ConfigureGatewayStore(t, app.Store, testutil.GatewayScenarioOpts{
		CompanyID:          provisioned.Company.ID,
		NewAPIWalletUserID: walletID,
		DepartmentID:       rootDept,
		Budget:             1000,
		UseRealWallet:      true,
		NewAPIMock:         mock,
	})
	rec = httptest.NewRecorder()
	app.Router.ServeHTTP(rec, gatewayRequest(fullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when wallet and budget ready, got %d body=%s", rec.Code, rec.Body.String())
	}
}
