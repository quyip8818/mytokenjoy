package gateway_test

import (
	"testing"

	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func TestDebugGatewayPrecheckError(t *testing.T) {
	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{
		WalletQuota: newapi.ToNewAPIUnits(100, nil, nil),
		Budget:      1000,
	})
	ctx := testutil.Ctx()
	mapping, err := scenario.Store.Relay().GetMappingByKeyHash(ctx, store.HashPlatformKey(scenario.FullKey))
	if err != nil || mapping == nil {
		t.Fatalf("mapping: %v", err)
	}
	company, err := scenario.Store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	precheck := relayfix.NewPrecheckService(scenario.Cfg, scenario.Store, relayfix.NewStubWallet(newapi.ToNewAPIUnits(100, nil, nil)))
	err = precheck.Run(ctx, domainrelay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   "gpt-4o",
	})
	if err != nil {
		t.Logf("precheck error: %v", err)
		t.Fatalf("precheck failed: %v", err)
	}
}
