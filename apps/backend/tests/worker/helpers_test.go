package worker_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	workerTF "github.com/tokenjoy/backend/tests/testutil/worker"
)

func newWorkerRunner(t *testing.T, stub *mock.StubAdminClient) (*worker.Runner, store.Store, *newapisync.NewAPISync) {
	t.Helper()
	runner, st, newAPISync, _ := workerTF.NewRunner(t, stub)
	return runner, st, newAPISync
}

func pendingNewAPISyncOutbox(st store.Store, kind string) int {
	return workerTF.PendingNewAPISyncOutbox(st, kind)
}
