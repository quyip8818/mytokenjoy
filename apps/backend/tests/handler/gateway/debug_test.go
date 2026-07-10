package gateway_test

import (
	"testing"

	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestDebugGatewayPrecheckError(t *testing.T) {
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		WalletQuota: newapi.ToNewAPIUnits(100, nil, nil),
		Budget:      1000,
	})
	ctx := testutil.Ctx()
	mapping, err := scenario.Store.PlatformKeyMappings().GetMappingByKeyHash(ctx, store.HashPlatformKey(scenario.FullKey))
	if err != nil || mapping == nil {
		t.Fatalf("mapping: %v", err)
	}
	company, err := scenario.Store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	precheck := gatewaytf.NewPrecheckService(scenario.Cfg, scenario.Store, gatewaytf.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domaingateway.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err != nil {
		t.Logf("precheck error: %v", err)
		t.Fatalf("precheck failed: %v", err)
	}
}
