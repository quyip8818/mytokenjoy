package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
)

func gatewayRequest(fullKey string) *http.Request {
	body, _ := json.Marshal(map[string]string{"model": "gpt-4o"})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+fullKey)
	return req
}

func TestGatewayRejectsInsufficientWallet(t *testing.T) {
	scenario := testutil.BuildGatewayScenario(t, testutil.GatewayScenarioOpts{
		WalletQuota: 0,
		Budget:      1000,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for empty wallet, got %d", rec.Code)
	}
}

func TestGatewayRejectsZeroBudget(t *testing.T) {
	units := newapi.ToNewAPIUnits(100, nil, nil)
	scenario := testutil.BuildGatewayScenario(t, testutil.GatewayScenarioOpts{
		WalletQuota: units,
		Budget:      0,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for zero budget, got %d", rec.Code)
	}
}

func TestGatewayAllowsWhenPrecheckPasses(t *testing.T) {
	units := newapi.ToNewAPIUnits(100, nil, nil)
	scenario := testutil.BuildGatewayScenario(t, testutil.GatewayScenarioOpts{
		WalletQuota: units,
		Budget:      1000,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when precheck passes, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGatewayRejectsExhaustedDepartmentBudget(t *testing.T) {
	units := newapi.ToNewAPIUnits(100, nil, nil)
	scenario := testutil.BuildGatewayScenario(t, testutil.GatewayScenarioOpts{
		WalletQuota: units,
		Budget:      100,
		Consumed:    100,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for exhausted budget, got %d", rec.Code)
	}
}

func TestGatewayRejectsInvalidAPIKey(t *testing.T) {
	scenario := testutil.BuildGatewayScenario(t, testutil.GatewayScenarioOpts{
		WalletQuota: newapi.ToNewAPIUnits(100, nil, nil),
		Budget:      1000,
	})
	body, _ := json.Marshal(map[string]string{"model": "gpt-4o"})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer sk-unknown-key")
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unknown key, got %d", rec.Code)
	}
}

func TestGatewayMountedOnRouter(t *testing.T) {
	mock := testutil.StartNewAPIMock(t)
	app := newTestApp(t, func(cfg *config.Config) {
		testutil.ApplySaaSConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.RelayGatewayEnabled = true
	})
	router := app.Router
	platformCookie := testutil.LoginPlatform(t, router)
	provisioned := testutil.ProvisionCompanyHTTP(t, router, platformCookie,
		"router-gw", "Router GW", "router-gw@example.com", "Router Admin", "securepass123")

	walletID := int64(0)
	if provisioned.Company.NewAPIWalletUserID != nil {
		walletID = *provisioned.Company.NewAPIWalletUserID
	}
	units := newapi.ToNewAPIUnits(100, nil, nil)
	mock.SetQuota(walletID, units)
	rootDept := fmt.Sprintf("dept-root-%d", provisioned.Company.ID)
	testutil.UpdateBudgetNodeHTTP(t, router, provisioned.MemberCookie, rootDept, 1000)

	fullKey := testutil.ConfigureGatewayStore(t, app.Store, testutil.GatewayScenarioOpts{
		CompanyID:          provisioned.Company.ID,
		NewAPIWalletUserID: walletID,
		DepartmentID:       rootDept,
		Budget:             1000,
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, gatewayRequest(fullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via router gateway, got %d body=%s", rec.Code, rec.Body.String())
	}
}
