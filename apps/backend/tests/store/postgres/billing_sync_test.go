//go:build testhook

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpsertOrderFromSync(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t)
	ctx := context.Background()

	orderID := uuid.Must(uuid.NewV7())
	order := store.RechargeOrder{
		ID:             orderID,
		CompanyID:      contract.DefaultCompanyID,
		Amount:         100.0,
		Currency:       "CNY",
		QuotaPerUnit:   500000,
		QuotaGranted:   50000000,
		Source:         "self",
		LotKind:        store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD-SYNC-001",
		PaymentMethod:  store.PaymentMethodAlipay,
		CreatedAt:      time.Now().UTC(),
	}

	// First insert succeeds.
	if err := st.Billing().UpsertOrderFromSync(ctx, order); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// Second insert (same ID) is idempotent — no error.
	if err := st.Billing().UpsertOrderFromSync(ctx, order); err != nil {
		t.Fatalf("second upsert (idempotent): %v", err)
	}

	// Verify order exists.
	got, err := st.Billing().GetRechargeOrder(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected order to exist")
	}
	if got.Amount != 100.0 {
		t.Fatalf("order amount: got %f want 100.0", got.Amount)
	}
}

func TestUpsertLotFromSync(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewFreshTestStore(t)
	ctx := context.Background()

	// Create prerequisite order.
	orderID := uuid.Must(uuid.NewV7())
	order := store.RechargeOrder{
		ID:           orderID,
		CompanyID:    contract.DefaultCompanyID,
		Amount:       200.0,
		Currency:     "CNY",
		QuotaPerUnit: 500000,
		QuotaGranted: 100000000,
		Source:       "platform",
		LotKind:      store.LotKindPaid,
		Status:       store.RechargeStatusConfirmed,
		CreatedAt:    time.Now().UTC(),
	}
	if err := st.Billing().UpsertOrderFromSync(ctx, order); err != nil {
		t.Fatal(err)
	}

	lotID := uuid.Must(uuid.NewV7())
	lot := store.RechargeLot{
		ID:              lotID,
		CompanyID:       contract.DefaultCompanyID,
		RechargeOrderID: orderID,
		BillingCurrency: "CNY",
		LotKind:         store.LotKindPaid,
		PaidAmount:      200.0,
		QuotaPerUnit:    500000,
		QuotaGranted:    100000000,
		QuotaRemaining:  80000000,
		Status:          store.LotStatusActive,
		CreatedAt:       time.Now().UTC(),
	}

	// First insert.
	if err := st.Billing().UpsertLotFromSync(ctx, lot); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// Verify lot.
	got, err := st.Billing().GetLotByID(ctx, lotID)
	if err != nil {
		t.Fatal(err)
	}
	if got.QuotaRemaining != 80000000 {
		t.Fatalf("lot quota_remaining: got %d want 80000000", got.QuotaRemaining)
	}

	// Update quota_remaining via second upsert (simulates consumption on SaaS).
	lot.QuotaRemaining = 50000000
	if err := st.Billing().UpsertLotFromSync(ctx, lot); err != nil {
		t.Fatalf("second upsert (update): %v", err)
	}

	got, err = st.Billing().GetLotByID(ctx, lotID)
	if err != nil {
		t.Fatal(err)
	}
	if got.QuotaRemaining != 50000000 {
		t.Fatalf("after update: got %d want 50000000", got.QuotaRemaining)
	}

	// Mark exhausted via upsert.
	lot.QuotaRemaining = 0
	lot.Status = store.LotStatusExhausted
	if err := st.Billing().UpsertLotFromSync(ctx, lot); err != nil {
		t.Fatalf("third upsert (exhaust): %v", err)
	}

	got, err = st.Billing().GetLotByID(ctx, lotID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.LotStatusExhausted {
		t.Fatalf("after exhaust: status got %q want %q", got.Status, store.LotStatusExhausted)
	}
}
