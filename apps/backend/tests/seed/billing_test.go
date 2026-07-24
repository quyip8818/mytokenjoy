package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/runtime"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApplyRechargeOrdersSeedsPostgres(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.CtxForCompany(contract.DefaultCompanyID)
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	// Template DB contains 5 recharge orders + 1 seed-lot order from ApplyUsageLedger.
	if len(orders) != 6 {
		t.Fatalf("expected 6 orders, got %d", len(orders))
	}
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	orders, err = st.Billing().ListRechargeOrders(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 6 {
		t.Fatalf("expected idempotent seed, got %d orders", len(orders))
	}
}
