package wallet_test

import (
	"testing"
	"time"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainwallet "github.com/tokenjoy/backend/internal/domain/wallet"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreditFromLotUpdatesWalletRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()

	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || co == nil {
		t.Fatal("expected default company")
	}
	before := co.WalletRemain

	key := "idem-wallet-credit"
	order := store.RechargeOrder{
		ID: "rch-wallet-1", CompanyID: contract.DefaultCompanyID, Amount: 50, Currency: "CNY",
		PointsPerUnit: ppu, PointsGranted: domainbilling.PointsGrantedFromAmount(50, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101130000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildPaidLot(order, "CNY", ppu)
	if err := domainwallet.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
		t.Fatal(err)
	}

	got, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || got == nil {
		t.Fatal("expected company after credit")
	}
	want := before + lot.PointsGranted
	if got.WalletRemain != want {
		t.Fatalf("wallet_remain: got %v want %v", got.WalletRemain, want)
	}
}

func TestConsumeLotsDecrementsWalletRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()
	grant := domainbilling.PointsGrantedFromAmount(10, ppu)

	order := store.RechargeOrder{
		ID: "rch-wallet-consume", CompanyID: contract.DefaultCompanyID, Amount: 10, Currency: "CNY",
		PointsPerUnit: ppu, PointsGranted: grant,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101140000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildPaidLot(order, "CNY", ppu)
	if err := domainwallet.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
		t.Fatal(err)
	}
	afterCredit, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || afterCredit == nil {
		t.Fatal("expected company after credit")
	}

	consume := grant / 4
	segments, err := domainwallet.ConsumeLots(ctx, st, contract.DefaultCompanyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) == 0 {
		t.Fatal("expected lot segments")
	}

	afterConsume, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || afterConsume == nil {
		t.Fatal("expected company after consume")
	}
	want := afterCredit.WalletRemain - consume
	if afterConsume.WalletRemain != want {
		t.Fatalf("wallet_remain: got %v want %v", afterConsume.WalletRemain, want)
	}
}
