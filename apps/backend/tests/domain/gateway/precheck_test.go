package gateway_test

import (
	"testing"

	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestPrecheckRejects(t *testing.T) {
	t.Parallel()

	for _, tc := range gatewaytf.RejectionCases() {
		if !tc.Precheck {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			fx := gatewaytf.NewPrecheckFixture(t, tc.Scenario)
			if err := fx.Run(tc.Model, false); err == nil {
				t.Fatalf("expected rejection for %s", tc.Name)
			}
		})
	}
}

func TestPrecheckRejectsInactivePlatformKey(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})

	keys, err := fx.Store.Keys().PlatformKeys(fx.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		keys[i].Status = "disabled"
	}
	if err := fx.Store.Keys().SetPlatformKeys(fx.Ctx, keys); err != nil {
		t.Fatal(err)
	}

	if err := fx.Run("gpt-4o", false); err == nil {
		t.Fatal("expected inactive platform key error")
	}
}

func TestPrecheckPassesWhenNewAPIUnavailable(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected precheck to pass without NewAPI wallet read, got %v", err)
	}
}

func TestPrecheckAllowsModelsListingWithoutModelField(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	if err := fx.Run("", true); err != nil {
		t.Fatalf("expected models listing precheck to pass, got %v", err)
	}
}

func TestPrecheckPassesRegardlessOfDeptConsumed(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected precheck to pass without budget consumed join, got %v", err)
	}
}
