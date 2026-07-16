package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestIngestEnqueueFailureRollsBackLedger(t *testing.T) {
	fix := newIngestFixture(t, withEnqueuer(mock.FailingEnqueuer{}))
	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(4102, 99))

	err := fix.Ingest.IngestByLogID(testutil.Ctx(), 4102, "webhook")
	if err == nil {
		t.Fatal("expected ingest to fail when enqueue fails")
	}

	ingested, err := testutil.HasLedgerLogID(fix.Store, 4102)
	if err != nil {
		t.Fatal(err)
	}
	if ingested {
		t.Fatal("expected no ledger row when enqueue fails inside transaction")
	}
	if riverfix.PendingWalletSyncCount(fix.Store, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no wallet_sync job after rollback")
	}
}
