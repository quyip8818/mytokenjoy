package gateway_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
	"github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestGatewayRejectionHTTPMapping(t *testing.T) {
	t.Parallel()

	for _, tc := range gatewaytf.RejectionCases() {
		if tc.WantHTTP == 0 {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			scenario := gatewaytf.BuildGatewayScenario(t, tc.Scenario)
			req := gatewaytf.GatewayRequestWithModel(scenario.FullKey, tc.Model)
			rec := httptest.NewRecorder()
			scenario.Gateway.ServeHTTP(rec, req)
			if rec.Code != tc.WantHTTP {
				t.Fatalf("expected %d for %s, got %d body=%s", tc.WantHTTP, tc.Name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGatewayProxiesDespiteZeroDeptBudget(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 0,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewaytf.GatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected gateway to pass without dept budget check, got %d", rec.Code)
	}
}

func TestGatewayProxiesDespiteExhaustedDepartmentBudget(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:   100,
		Consumed: 100,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewaytf.GatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected gateway to pass without consumed join, got %d", rec.Code)
	}
}

func TestGatewayProxiesFullV1Path(t *testing.T) {
	t.Parallel()
	var upstreamPath atomic.Value
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamPath.Store(r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:          1000,
		ProxyBackendURL: backend.URL,
	})
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, gatewaytf.GatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	path, _ := upstreamPath.Load().(string)
	if path != "/v1/chat/completions" {
		t.Fatalf("expected upstream path /v1/chat/completions, got %q", path)
	}
}

func TestGatewayRejectsSubpath(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 1000,
	})
	req := gatewaytf.GatewayRequestWithModel(scenario.FullKey, "deepseek-v4-pro")
	req.URL.Path = "/v1/chat/completions/evil"
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for subpath, got %d", rec.Code)
	}
}

func TestGatewayRejectsOversizedBody(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 1000,
	})
	oversized := make([]byte, (4<<20)+1)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(oversized))
	req.Header.Set("Authorization", "Bearer "+scenario.FullKey)
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d", rec.Code)
	}
}

func TestGatewayMountedOnRouter(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.GatewayEnabled = true
	})
	router := app.Router
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Router GW", "router-gw@example.com", "Router Admin", "securepass123")

	walletID := int64(0)
	if provisioned.Company.NewAPIWalletCompanyID != nil {
		walletID = *provisioned.Company.NewAPIWalletCompanyID
	}
	rootDept := uuid.Nil
	if provisioned.Company.RootDeptID != nil {
		rootDept = *provisioned.Company.RootDeptID
	}
	ctx := testutil.CtxForCompany(provisioned.Company.ID)
	if err := app.Store.Company().SetWalletQuotaRemain(ctx, provisioned.Company.ID, 100000, nil); err != nil {
		t.Fatal(err)
	}

	fullKey := gatewaytf.ConfigureGatewayStore(t, app.Config, app.Store, gatewaytf.GatewayScenarioOpts{
		CompanyID:             provisioned.Company.ID,
		NewAPIWalletCompanyID: walletID,
		DepartmentID:          rootDept,
		Budget:                1000,
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, gatewaytf.GatewayRequest(fullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via router gateway, got %d body=%s", rec.Code, rec.Body.String())
	}
}
