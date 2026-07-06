package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApplyRechargeOrdersSeedsMemoryStore(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.CtxForCompany(seed.DefaultCompanyID)
	if err := seed.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 5 {
		t.Fatalf("expected 5 orders, got %d", len(orders))
	}
	if err := seed.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	orders, err = st.Billing().ListRechargeOrders(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 5 {
		t.Fatalf("expected idempotent seed, got %d orders", len(orders))
	}
}
