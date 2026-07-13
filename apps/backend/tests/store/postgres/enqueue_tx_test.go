package postgres_test

import (
	"fmt"
	"testing"

	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestInsertInTxRollsBackWithStoreTransaction(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)

	err := st.WithTx(ctx, func(txSt store.Store) error {
		tx, ok := txSt.(store.Tx)
		if !ok {
			t.Fatal("expected transactional store")
		}
		if err := jobs.InsertWalletSync(ctx, enqueuer, tx, contract.DefaultCompanyID); err != nil {
			return err
		}
		return fmt.Errorf("force rollback")
	})
	if err == nil {
		t.Fatal("expected forced rollback error")
	}
	if riverfix.PendingWalletSyncCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no wallet_sync job after transaction rollback")
	}
}

func TestInsertInTxCommitsWithStoreTransaction(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)

	if err := st.WithTx(ctx, func(txSt store.Store) error {
		tx, ok := txSt.(store.Tx)
		if !ok {
			t.Fatal("expected transactional store")
		}
		return jobs.InsertWalletSync(ctx, enqueuer, tx, contract.DefaultCompanyID)
	}); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingWalletSyncCount(st, contract.DefaultCompanyID) != 1 {
		t.Fatal("expected wallet_sync job after transaction commit")
	}
}
