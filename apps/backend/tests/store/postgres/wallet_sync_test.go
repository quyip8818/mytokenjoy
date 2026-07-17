package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestEnqueueWalletSyncDebouncesAndSlides(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	companyID := contract.DefaultCompanyID
	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)

	if err := jobs.InsertWalletSync(ctx, enqueuer, nil, companyID); err != nil {
		t.Fatal(err)
	}
	firstCount := walletSyncPendingCount(t, st, companyID)
	if firstCount != 1 {
		t.Fatalf("expected 1 pending wallet_sync job, got %d", firstCount)
	}

	if err := jobs.InsertWalletSync(ctx, enqueuer, nil, companyID); err != nil {
		t.Fatal(err)
	}
	secondCount := walletSyncPendingCount(t, st, companyID)
	if secondCount != 1 {
		t.Fatalf("expected unique wallet_sync debounce to keep one job, got %d", secondCount)
	}
}

func walletSyncPendingCount(t *testing.T, st store.Store, companyID uuid.UUID) int {
	t.Helper()
	pool := postgres.MainPool(st)
	var count int
	err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND (args->>'company_id')::uuid = $2
	`, jobs.KindWalletSync, companyID).Scan(&count)
	if err != nil {
		t.Fatalf("query wallet_sync jobs: %v", err)
	}
	return count
}
