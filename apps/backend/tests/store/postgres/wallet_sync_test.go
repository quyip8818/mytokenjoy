package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestEnqueueWalletSyncDebouncesAndSlides(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	companyID := contract.DefaultCompanyID

	before := time.Now().UTC()
	if err := st.AsyncJobs().EnqueueWalletSync(ctx, companyID); err != nil {
		t.Fatal(err)
	}
	first := walletSyncNextRetry(t, st, companyID)
	if !first.After(before.Add(time.Duration(common.WalletSyncDebounceSecs-1) * time.Second)) {
		t.Fatalf("expected debounced next_retry, got %v (before=%v)", first, before)
	}

	time.Sleep(10 * time.Millisecond)
	if err := st.AsyncJobs().EnqueueWalletSync(ctx, companyID); err != nil {
		t.Fatal(err)
	}
	second := walletSyncNextRetry(t, st, companyID)
	if second.Before(first) {
		t.Fatalf("expected sliding debounce, second=%v first=%v", second, first)
	}

	pending, err := st.AsyncJobs().HasPendingWalletSync(ctx, companyID)
	if err != nil {
		t.Fatal(err)
	}
	if !pending {
		t.Fatal("expected pending wallet_sync job")
	}
}

func walletSyncNextRetry(t *testing.T, st store.Store, companyID int64) time.Time {
	t.Helper()
	pool := postgres.MainPool(st)
	var nextRetry time.Time
	err := pool.QueryRow(context.Background(), `
		SELECT next_retry FROM async_jobs
		WHERE company_id = $1 AND channel = 'wallet_sync' AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, companyID).Scan(&nextRetry)
	if err != nil {
		t.Fatalf("query wallet_sync next_retry: %v", err)
	}
	return nextRetry
}
