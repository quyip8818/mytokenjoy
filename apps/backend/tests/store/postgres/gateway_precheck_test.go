package postgres_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestLoadPrecheckContextReturnsWalletAndRouting(t *testing.T) {
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
	if row.WalletRemainQuota != 50000 {
		t.Fatalf("expected wallet remain 50000, got %v", row.WalletRemainQuota)
	}
	if row.KeyStatus != "active" {
		t.Fatalf("expected active key, got %s", row.KeyStatus)
	}
}

func TestLoadPrecheckContextReturnsNilForUnknownKey(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	row, err := st.GatewayPrecheck().LoadPrecheckContext(ctx, store.HashPlatformKey("sk-unknown"))
	if err != nil {
		t.Fatal(err)
	}
	if row != nil {
		t.Fatal("expected nil row for unknown key hash")
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
	for _, want := range []string{"deepseek-v4-pro", "deepseek-v4-flash", "test-model"} {
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
