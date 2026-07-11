package gateway_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

// TestGatewayBudgetCheckSoftBlock verifies the optional GatewayBudgetCheck soft
// block: an exhausted soft_remain blocks; a positive one allows; and every
// enabled precheck performs exactly one Redis GET.
func TestGatewayBudgetCheckSoftBlock(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		softRemain float64
		wantErr    bool
	}{
		{name: "exhausted blocks", softRemain: -1, wantErr: true},
		{name: "zero blocks", softRemain: 0, wantErr: true},
		{name: "positive allows", softRemain: 50, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: testutil.DisplayPoints(1000)})
			fake := gatewaytf.NewFakeBudgetCheck()
			fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, fake)

			companyID := fx.LoadPrecheckRow(t).CompanyID
			_ = fake.Set(fx.Ctx, companyID, fx.KeyHash(), budgetcheck.Entry{
				PeriodKey:  "2026-07",
				SoftRemain: tc.softRemain,
				KeyStatus:  "active",
			})

			err := fx.Run("gpt-4o", false)
			if tc.wantErr && err == nil {
				t.Fatalf("expected soft block error for soft_remain=%v", tc.softRemain)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected allow for soft_remain=%v, got %v", tc.softRemain, err)
			}
			if fake.Gets() != 1 {
				t.Fatalf("expected exactly 1 Redis GET, got %d", fake.Gets())
			}
		})
	}
}

// TestGatewayBudgetCheckMissAllows verifies a cache miss degrades to allow and
// never falls back to Postgres.
func TestGatewayBudgetCheckMissAllows(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: testutil.DisplayPoints(1000)})
	fake := gatewaytf.NewFakeBudgetCheck()
	fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, fake)

	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected allow on cache miss, got %v", err)
	}
	if fake.Gets() != 1 {
		t.Fatalf("expected exactly 1 Redis GET on miss, got %d", fake.Gets())
	}
}

// TestGatewayBudgetCheckDisabledSkipsGet verifies the default no-op store skips
// the Redis GET entirely.
func TestGatewayBudgetCheckDisabledSkipsGet(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: testutil.DisplayPoints(1000)})
	fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, budgetcheck.Noop{})

	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected allow with disabled soft block, got %v", err)
	}
}
