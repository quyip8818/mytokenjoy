package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestIngestEnqueuesDashboardProjectAndWalletSync(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	_, st, ingest := riverfix.NewIngestRuntime(t, stub)

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4101, 99))

	if err := ingest.IngestByLogID(testutil.Ctx(), 4101, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	if riverfix.PendingDashboardProjectCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected dashboard_project job after ingest")
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no rebalance jobs directly from ingest")
	}
	if riverfix.PendingOverrunCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no overrun jobs directly from ingest")
	}
	if riverfix.PendingWalletSyncCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected wallet_sync job after ingest")
	}
}

func TestIngestEnqueueFailureRollsBackLedger(t *testing.T) {
	fix := newIngestFixture(t, withEnqueuer(mock.FailingEnqueuer{}))
	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(4102, 99))

	err := fix.Ingest.IngestByLogID(testutil.Ctx(), 4102, types.SourceWebhook)
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
