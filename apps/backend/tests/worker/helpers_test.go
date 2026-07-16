package worker_test

import (
	"context"
	"testing"

	newapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type workerFixture struct {
	rt         *riverfix.TestRuntime
	st         store.Store
	newAPISync *newapisync.NewAPISync
	ctx        context.Context
}

func newWorkerFixture(t *testing.T, stub *mock.StubAdminClient) workerFixture {
	t.Helper()
	rt, st := riverfix.NewRuntime(t, stub)
	nas := rt.Registry.MustNewAPISync()
	ctx := context.Background()
	t.Cleanup(func() {
		if rt != nil {
			rt.Stop(t, ctx)
		}
	})
	return workerFixture{
		rt:         rt,
		st:         st,
		newAPISync: nas,
		ctx:        ctx,
	}
}

func (f workerFixture) runRiver(t *testing.T) {
	t.Helper()
	f.rt.RunOnce(t, f.ctx)
}
