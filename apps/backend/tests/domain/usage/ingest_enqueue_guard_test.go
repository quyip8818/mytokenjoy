package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestIngestSkipsRebalanceAndOverrunWhenNewAPIDisabled(t *testing.T) {
	t.Parallel()
	_, st, ingest := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(92001, 99))

	if err := ingest.IngestByLogID(ctx, 92001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if testutil.PendingRebalanceCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no pending rebalance jobs when NewAPI disabled")
	}
	if testutil.PendingOverrunCount(st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no pending overrun jobs when NewAPI disabled")
	}
}
