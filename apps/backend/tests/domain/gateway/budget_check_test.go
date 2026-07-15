package gateway_test

import (
	"testing"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestPrecheckBlocksOnCombinedKeyRemain(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
	budgetfix.SetCombinedKeyRemain(t, fx.Store, fx.LoadPrecheckRow(t).PlatformKeyID, 0)
	if err := fx.Run("gpt-4o", false); err == nil {
		t.Fatal("expected PG soft summary block")
	}
}

func TestPrecheckAllowsNullCombinedKeyRemain(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected allow when soft summary NULL, got %v", err)
	}
}

func TestGatewayBudgetCheckCombinedKeyBlock(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		remain    float64
		version   int64
		pgVersion int64
		wantErr   bool
	}{
		{name: "exhausted blocks", remain: -1, version: 1, pgVersion: 1, wantErr: true},
		{name: "zero blocks", remain: 0, version: 1, pgVersion: 1, wantErr: true},
		{name: "positive allows", remain: 50, version: 1, pgVersion: 1, wantErr: false},
		{name: "stale version allows", remain: 0, version: 1, pgVersion: 2, wantErr: false},
		{name: "no pg version ignores redis", remain: 0, version: 1, pgVersion: 0, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
			fake := gatewaytf.NewFakeBudgetCheck()
			fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, fake)

			row := fx.LoadPrecheckRow(t)
			if tc.pgVersion > 0 {
				budgetfix.SetCombinedKeyRemain(t, fx.Store, row.PlatformKeyID, 1)
				for i := int64(1); i < tc.pgVersion; i++ {
					budgetfix.SetCombinedKeyRemain(t, fx.Store, row.PlatformKeyID, 1)
				}
			}
			row = fx.LoadPrecheckRow(t)

			_ = fake.Set(fx.Ctx, row.CompanyID, fx.KeyHash(), domainbudget.CombinedKeyEntry{
				Remain:  tc.remain,
				Version: tc.version,
			})

			err := fx.Run("gpt-4o", false)
			if tc.wantErr && err == nil {
				t.Fatalf("expected soft block error for soft_remain=%v version=%d pgVersion=%d", tc.remain, tc.version, row.CombinedKeyRemainVersion)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected allow for soft_remain=%v version=%d pgVersion=%d, got %v", tc.remain, tc.version, row.CombinedKeyRemainVersion, err)
			}
			if fake.Gets() != 1 {
				t.Fatalf("expected exactly 1 Redis GET, got %d", fake.Gets())
			}
		})
	}
}

func TestGatewayBudgetCheckMissAllows(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
	fake := gatewaytf.NewFakeBudgetCheck()
	fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, fake)

	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected allow on cache miss, got %v", err)
	}
	if fake.Gets() != 1 {
		t.Fatalf("expected exactly 1 Redis GET on miss, got %d", fake.Gets())
	}
}

func TestGatewayBudgetCheckDisabledSkipsGet(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: budgetfix.DisplayPoints(1000)})
	fx.Precheck = gatewaytf.NewPrecheckService(fx.Cfg, fx.Store, domainbudget.NoopCombinedKeyCache)

	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected allow with disabled soft block, got %v", err)
	}
}
