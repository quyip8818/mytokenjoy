package worker_test

import (
	"testing"

	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func newWorkerRunner(t *testing.T, stub *mock.StubAdminClient) (*worker.Runner, store.Store, *relay.TokenLifecycle) {
	t.Helper()
	runner, st, lifecycle, _ := workerfix.NewRunner(t, stub)
	return runner, st, lifecycle
}

func pendingRelayOutbox(st store.Store, kind string) int {
	return workerfix.PendingRelayOutbox(st, kind)
}
