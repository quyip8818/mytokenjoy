//go:build testhook

package gateway_test

import (
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func setCompanyType(t *testing.T, st store.Store, companyType string) {
	t.Helper()
	if err := st.Company().UpdateType(testutil.Ctx(), contract.DefaultCompanyID, companyType); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayAllowsTestModelForTrialCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeTrial)

	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "test-model")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected test-model allowed for trial, got 403 body=%s", rec.Body.String())
	}
}

func TestGatewayAllowsTestModelForDemoCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeDemo)

	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "test-model")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected test-model allowed for demo, got 403 body=%s", rec.Body.String())
	}
}

func TestGatewayRejectsTestModelForStandardCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000, CompanyType: store.CompanyTypeStandard})

	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "test-model")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("expected test-model rejected for standard, got %d", rec.Code)
	}
}

func TestGatewayAllowsRealModelForTrialCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeTrial)

	// trial accounts can use real models (subject to routing rules, not company type).
	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "deepseek-v4-pro")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected real model allowed for trial, got 403 body=%s", rec.Body.String())
	}
}

func TestGatewayAllowsRealModelForDemoCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeDemo)

	// demo accounts can use real models (subject to routing rules, not company type).
	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "deepseek-v4-pro")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected real model allowed for demo, got 403 body=%s", rec.Body.String())
	}
}

func TestGatewayAllowsTestModelForTestingCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeTesting)

	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "test-model")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected test-model allowed for testing, got 403 body=%s", rec.Body.String())
	}
}

func TestGatewayAllowsRealModelForTestingCompany(t *testing.T) {
	t.Parallel()
	sc := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	setCompanyType(t, sc.Store, store.CompanyTypeTesting)

	req := gatewaytf.GatewayRequestWithModel(sc.FullKey, "deepseek-v4-pro")
	rec := httptest.NewRecorder()
	sc.Gateway.ServeHTTP(rec, req)
	if rec.Code == 403 {
		t.Fatalf("expected real model allowed for testing, got 403 body=%s", rec.Body.String())
	}
}
