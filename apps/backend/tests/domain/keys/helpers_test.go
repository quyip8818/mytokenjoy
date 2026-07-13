package keys_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func testSyncEnqueuer(t *testing.T, cfg config.Config, st store.Store) ports.SyncJobEnqueuer {
	t.Helper()
	return app.NewNewAPISyncEnqueuer(riverfix.NewInsertOnlyEnqueuer(t, cfg, st))
}

func newKeysService(t *testing.T) (domainkeys.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	newAPISync := newapisync.New(cfg, st, nil, nil, policy.NewChannelPolicy(cfg), testSyncEnqueuer(t, cfg, st))
	return domainkeys.NewService(cfg, st, newAPISync, common.NewDelayer(false)), st
}

func newKeysServiceWithNewAPI(t *testing.T) (domainkeys.Service, store.Store, *mock.StubAdminClient) {
	t.Helper()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 883, Key: "sk-test-key", RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	newAPISync := newapisync.New(cfg, st, newapi.NewAdminPortAdapter(stub), nil, policy.NewChannelPolicy(cfg), testSyncEnqueuer(t, cfg, st))
	return domainkeys.NewService(cfg, st, newAPISync, common.NewDelayer(false)), st, stub
}

func newNewAPISync(t *testing.T, stub *mock.StubAdminClient) (*newapisync.NewAPISync, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	return newapisync.New(cfg, st, newapi.NewAdminPortAdapter(stub), nil, policy.NewChannelPolicy(cfg), testSyncEnqueuer(t, cfg, st)), st
}

func findApproval(st store.Store, id string) *types.KeyApproval {
	approvals, err := st.Keys().Approvals(testutil.Ctx())
	if err != nil {
		return nil
	}
	for _, a := range approvals {
		if a.ID == id {
			copy := a
			return &copy
		}
	}
	return nil
}
