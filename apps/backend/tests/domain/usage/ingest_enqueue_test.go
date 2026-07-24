package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func TestIngestEnqueueFailureRollsBackLedger(t *testing.T) {
	fix := newIngestFixture(t, withEnqueuer(mock.FailingEnqueuer{}))
	// Set combined key remain to negative so overrun enqueue is triggered.
	budgetfix.SetCombinedKeyRemain(t, fix.Store, newapisynctf.DefaultMappingOpts().PlatformKeyID, -1)
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
}
