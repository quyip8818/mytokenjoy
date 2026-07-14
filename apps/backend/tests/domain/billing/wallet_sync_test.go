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
	"github.com/tokenjoy/backend/internal/store"
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
		BillingCurrency: "CNY", CreatedAt: now, UpdatedAt: now,
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
		NewAPIWalletUserID: &zero, BillingCurrency: "CNY", CreatedAt: now, UpdatedAt: now,
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
