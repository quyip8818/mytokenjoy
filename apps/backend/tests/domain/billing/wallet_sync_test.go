package billing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestSyncCompanyWalletReturnsErrWalletNotConfiguredWhenMissingWallet(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, newapi.NewAdminPortAdapter(client), company.NewWalletService(cfg, client), domainbilling.NoopJobEnqueuer)

	const companyID int64 = 888_001
	ctx := context.Background()
	now := time.Now().UTC()
	if err := st.Company().Create(ctx, store.Company{
		ID: companyID, Slug: "no-wallet", Name: "No Wallet", Status: store.CompanyStatusActive,
		BillingCurrency: common.DefaultBillingCurrency, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	err := svc.SyncCompanyWallet(company.WithContext(ctx, company.Context{CompanyID: companyID}), companyID)
	if !errors.Is(err, domainbilling.ErrWalletNotConfigured) {
		t.Fatalf("expected ErrWalletNotConfigured, got %v", err)
	}
}

func TestSyncCompanyWalletReturnsErrWalletNotConfiguredWhenWalletIDZero(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, newapi.NewAdminPortAdapter(client), company.NewWalletService(cfg, client), domainbilling.NoopJobEnqueuer)

	const companyID int64 = 888_002
	ctx := context.Background()
	now := time.Now().UTC()
	zero := int64(0)
	if err := st.Company().Create(ctx, store.Company{
		ID: companyID, Slug: "zero-wallet", Name: "Zero Wallet", Status: store.CompanyStatusActive,
		NewAPIWalletUserID: &zero, BillingCurrency: common.DefaultBillingCurrency, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	err := svc.SyncCompanyWallet(company.WithContext(ctx, company.Context{CompanyID: companyID}), companyID)
	if !errors.Is(err, domainbilling.ErrWalletNotConfigured) {
		t.Fatalf("expected ErrWalletNotConfigured, got %v", err)
	}
}

func TestSyncCompanyWalletPropagatesStoreErrors(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, newapi.NewAdminPortAdapter(client), company.NewWalletService(cfg, client), domainbilling.NoopJobEnqueuer)

	err := svc.SyncCompanyWallet(context.Background(), 9_999_999_999)
	if err == nil {
		t.Fatal("expected error for missing company")
	}
	if errors.Is(err, domainbilling.ErrWalletNotConfigured) {
		t.Fatal("store lookup failure must not be mapped to ErrWalletNotConfigured")
	}
}

func TestSyncCompanyWalletIgnoresStaleLowCache(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	const walletUserID int64 = 777
	var quota int64
	var lastTopUp newapi.TopUpRequest
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, id int64) (int64, error) {
			if id != walletUserID {
				t.Fatalf("unexpected wallet user %d", id)
			}
			return quota, nil
		},
		TopUpFn: func(_ context.Context, req newapi.TopUpRequest) error {
			lastTopUp = req
			quota += req.Quota
			return nil
		},
	}
	wallet := company.NewWalletService(cfg, client)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, newapi.NewAdminPortAdapter(client), wallet, domainbilling.NoopJobEnqueuer)

	ctx := context.Background()
	coID := contract.DefaultCompanyID
	if err := st.Company().UpdateNewAPIWalletUserID(ctx, coID, walletUserID); err != nil {
		t.Fatal(err)
	}
	// Poison cache with a low reading while NewAPI already holds a near-max quota.
	quota = 0
	if _, err := wallet.AvailableNewAPIUnits(ctx, walletUserID); err != nil {
		t.Fatal(err)
	}
	quota = 4_980_468_515_629_721_982
	if err := svc.SyncCompanyWallet(company.WithContext(ctx, company.Context{CompanyID: coID}), coID); err != nil {
		t.Fatal(err)
	}
	if lastTopUp.UserID != walletUserID {
		t.Fatalf("unexpected TopUp user %+v", lastTopUp)
	}
	if lastTopUp.Quota >= 0 {
		t.Fatalf("expected subtract TopUp from stale-cache guard, got %+v (calls=%d)", lastTopUp, client.TopUpCalls)
	}
}
