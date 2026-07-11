package gateway_test

import (
	"testing"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestPrecheckRejects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts gatewaytf.GatewayScenarioOpts
		cfg  []testutil.ConfigOption
		run  func(t *testing.T, fx gatewaytf.PrecheckFixture)
	}{
		{
			name: "zero budget",
			opts: gatewaytf.GatewayScenarioOpts{Budget: 0},
			run: func(t *testing.T, fx gatewaytf.PrecheckFixture) {
				if err := fx.Run("gpt-4o", false); err == nil {
					t.Fatal("expected budget exceeded error")
				}
			},
		},
		{
			name: "model not in whitelist",
			opts: gatewaytf.GatewayScenarioOpts{Budget: 1000},
			run: func(t *testing.T, fx gatewaytf.PrecheckFixture) {
				if err := fx.Run("unknown-model", false); err == nil {
					t.Fatal("expected model not allowed error")
				}
			},
		},
		{
			name: "suspended company",
			opts: gatewaytf.GatewayScenarioOpts{Budget: 1000, CompanyStatus: "suspended"},
			run: func(t *testing.T, fx gatewaytf.PrecheckFixture) {
				if err := fx.Run("gpt-4o", false); err == nil {
					t.Fatal("expected suspended company error")
				}
			},
		},
		{
			name: "insufficient wallet",
			opts: gatewaytf.GatewayScenarioOpts{
				Budget:             1000,
				WalletBalancePoint: testutil.Float64Ptr(0),
			},
			run: func(t *testing.T, fx gatewaytf.PrecheckFixture) {
				if err := fx.Run("gpt-4o", false); err == nil {
					t.Fatal("expected insufficient wallet error")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := gatewaytf.NewPrecheckFixture(t, tc.opts, tc.cfg...)
			tc.run(t, fx)
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

func TestPrecheckIgnoresNewAPIKeyRemainQuota(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{
		Budget:      1000,
		RemainQuota: 0,
	})
	if err := fx.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected precheck to pass when NewAPIKeyRemainQuota is zero, got %v", err)
	}
}

func TestPrecheckAllowsModelsListingWithoutModelField(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	if err := fx.Run("", true); err != nil {
		t.Fatalf("expected models listing precheck to pass, got %v", err)
	}
}

func TestPrecheckUsesClockAnchorForPeriodKey(t *testing.T) {
	t.Parallel()
	fxJune := gatewaytf.NewPrecheckFixture(t,
		gatewaytf.GatewayScenarioOpts{Budget: testutil.DisplayPoints(1000)},
		testutil.WithClockAnchor("2026-06-19"),
	)

	junePeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, fxJune.Cfg.Clock()).String()
	testutil.SetSnapshotConsumedAtPeriod(t, fxJune.Store, store.SnapshotAxisOrgNode, contract.IDDept3, junePeriod, testutil.DisplayPoints(1000))
	if err := fxJune.Run("gpt-4o", false); err == nil {
		t.Fatal("expected budget exceeded when clock anchors June period")
	}

	fxJuly := gatewaytf.NewPrecheckFixture(t,
		gatewaytf.GatewayScenarioOpts{Budget: testutil.DisplayPoints(100000)},
		testutil.WithClockAnchor("2026-07-15"),
	)
	if err := fxJuly.Run("gpt-4o", false); err != nil {
		t.Fatalf("expected precheck to pass for July period with no consumption, got %v", err)
	}
}
