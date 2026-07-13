package newapisync_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func newSyncWithStubAndCfg(t *testing.T, stub *mock.StubAdminClient, opts ...testutil.ConfigOption) (*newapisync.NewAPISync, store.Store) {
	t.Helper()
	if len(opts) > 0 {
		return newapisynctf.NewLocalTestService(t, stub, opts...)
	}
	sync, _, st := newapisynctf.NewTestService(t, newapisynctf.TestServiceOpts{Stub: stub})
	return sync, st
}

func newSyncWithStub(t *testing.T, stub *mock.StubAdminClient) (*newapisync.NewAPISync, store.Store) {
	return newSyncWithStubAndCfg(t, stub)
}
