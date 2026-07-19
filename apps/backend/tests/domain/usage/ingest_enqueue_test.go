package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
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
}
