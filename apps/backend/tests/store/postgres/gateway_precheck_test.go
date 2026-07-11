package postgres_test

import (
	"testing"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestLoadPrecheckContextReturnsMappingAndWallet(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{
		Budget:             1000,
		WalletBalancePoint: testutil.Float64Ptr(50000),
	})

	row := fx.LoadPrecheckRow(t)
	if row == nil {
		t.Fatal("expected precheck context row")
	}
	if row.CompanyID != contract.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", contract.DefaultCompanyID, row.CompanyID)
	}
	if row.BalancePoint != 50000 {
		t.Fatalf("expected balance 50000, got %v", row.BalancePoint)
	}
	if row.DepartmentID != contract.IDDept3 {
		t.Fatalf("expected department %s, got %s", contract.IDDept3, row.DepartmentID)
	}
	if !row.DeptFound {
		t.Fatal("expected department found")
	}
	if row.KeyStatus != "active" {
		t.Fatalf("expected active key, got %s", row.KeyStatus)
	}
}

func TestLoadPrecheckContextReturnsNilForUnknownKey(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	row, err := st.GatewayPrecheck().LoadPrecheckContext(ctx, store.HashPlatformKey("sk-unknown"), cfg.Clock().Now())
	if err != nil {
		t.Fatal(err)
	}
	if row != nil {
		t.Fatal("expected nil row for unknown key hash")
	}
}

func TestLoadPrecheckContextPeriodKeyAlignment(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t,
		gatewaytf.GatewayScenarioOpts{
			Budget:   testutil.DisplayPoints(1000),
			Consumed: testutil.DisplayPoints(250),
		},
		testutil.WithClockAnchor("2026-06-19"),
	)

	periodKey := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, fx.Cfg.Clock()).String()
	row := fx.LoadPrecheckRow(t)
	if row.PeriodKey != periodKey {
		t.Fatalf("expected period %q, got %q", periodKey, row.PeriodKey)
	}
	if row.DeptConsumed != testutil.DisplayPoints(250) {
		t.Fatalf("expected dept consumed %v, got %v", testutil.DisplayPoints(250), row.DeptConsumed)
	}
}

func TestLoadPrecheckContextAllowlistTypes(t *testing.T) {
	t.Parallel()
	fx := gatewaytf.NewPrecheckFixture(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	row := fx.LoadPrecheckRow(t)

	if row.PlatformKeyID != contract.IDPlatformKey1 {
		t.Fatalf("expected platform key %s, got %s", contract.IDPlatformKey1, row.PlatformKeyID)
	}
	if !row.HasAllowlist {
		t.Fatal("expected platform key allowlist to be present")
	}
	for _, want := range []string{"claude-sonnet-4-6", "gpt-4o"} {
		if !contains(row.AllowlistTypes, want) {
			t.Fatalf("expected allowlist to include %q, got %v", want, row.AllowlistTypes)
		}
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
