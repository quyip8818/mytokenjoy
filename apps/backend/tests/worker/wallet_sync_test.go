package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestWalletSyncWorkerTopUpOnDrift(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
	}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	const walletID int64 = 501
	if err := fix.st.Company().UpdateNewAPIWalletUserID(ctx, contract.DefaultCompanyID, walletID); err != nil {
		t.Fatal(err)
	}
	if err := fix.rt.Registry.BillingSvc.PlatformRecharge(ctx, contract.DefaultCompanyID, 100, "wallet-sync-test"); err != nil {
		t.Fatal(err)
	}

	if err := jobs.InsertWalletSync(ctx, fix.rt.Enqueuer, nil, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	fix.runRiver(t)

	if stub.TopUpCalls == 0 {
		t.Fatal("expected wallet_sync worker to call TopUp on positive drift")
	}
	if testutil.PendingWalletSyncCount(fix.st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected wallet_sync job drained")
	}
}

func TestWalletSyncWorkerCancelsWhenWalletNotConfigured(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	const companyID int64 = 888_002
	now := time.Now().UTC()
	if err := fix.st.Company().Create(ctx, store.Company{
		ID: companyID, Slug: "wallet-sync-cancel", Name: "No Wallet Co",
		Status: store.CompanyStatusActive, BillingCurrency: "CNY", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := jobs.InsertWalletSync(ctx, fix.rt.Enqueuer, nil, companyID); err != nil {
		t.Fatal(err)
	}
	fix.runRiver(t)

	pool := postgres.MainPool(fix.st)
	var state string
	if err := pool.QueryRow(ctx, `
		SELECT state::text FROM river_job
		WHERE kind = $1 AND (args->>'company_id')::bigint = $2
		ORDER BY id DESC LIMIT 1
	`, jobs.KindWalletSync, companyID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "cancelled" && state != "discarded" {
		t.Fatalf("expected cancelled wallet_sync job, got %q", state)
	}
}
