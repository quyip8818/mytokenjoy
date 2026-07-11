package usage_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestIngestEnqueuesBudgetProjectAndWalletSync(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	_, st, ingest := workerfix.NewRuntime(t, stub)

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4101, 99))

	if err := ingest.IngestByLogID(testutil.Ctx(), 4101, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	if testutil.PendingBudgetProjectCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected budget_project job after ingest")
	}
	if testutil.PendingRebalanceCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no rebalance jobs directly from ingest")
	}
	if testutil.PendingOverrunCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no overrun jobs directly from ingest")
	}
	if testutil.PendingWalletSyncCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected wallet_sync job after ingest")
	}
}

func TestIngestEnqueueFailureRollsBackLedger(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithIngestEnabled(true),
	)
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4102, 99))

	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, holder, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(stub))
	if err != nil {
		t.Fatal(err)
	}
	holder.Set(failingEnqueuer{})
	ingest := reg.MustIngestService()

	err = ingest.IngestByLogID(testutil.Ctx(), 4102, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected ingest to fail when enqueue fails")
	}

	ingested, err := testutil.HasLedgerLogID(st, 4102)
	if err != nil {
		t.Fatal(err)
	}
	if ingested {
		t.Fatal("expected no ledger row when enqueue fails inside transaction")
	}
	if testutil.PendingWalletSyncCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no wallet_sync job after rollback")
	}
	if testutil.PendingBudgetProjectCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no budget_project job after rollback")
	}
}

type failingEnqueuer struct{}

func (failingEnqueuer) Insert(context.Context, river.JobArgs, *river.InsertOpts) error {
	return fmt.Errorf("enqueue failed")
}

func (failingEnqueuer) InsertInTx(context.Context, store.Tx, river.JobArgs, *river.InsertOpts) error {
	return fmt.Errorf("enqueue failed")
}

var _ jobs.Enqueuer = failingEnqueuer{}
