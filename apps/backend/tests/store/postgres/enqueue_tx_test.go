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
		if err := jobs.InsertRebalance(ctx, enqueuer, tx, contract.DefaultCompanyID, store.RebalanceAxisCompany, contract.DefaultCompanyID); err != nil {
			return err
		}
		return fmt.Errorf("force rollback")
	})
	if err == nil {
		t.Fatal("expected forced rollback error")
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no rebalance job after transaction rollback")
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
		return jobs.InsertRebalance(ctx, enqueuer, tx, contract.DefaultCompanyID, store.RebalanceAxisCompany, contract.DefaultCompanyID)
	}); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) != 1 {
		t.Fatal("expected rebalance job after transaction commit")
	}
}
