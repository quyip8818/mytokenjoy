package keys_test

import (
	"testing"

	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newKeysService(t *testing.T) (domainkeys.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	newAPISync := newapisync.New(cfg, st, nil, nil, newapisync.NewChannelPolicy(cfg))
	return domainkeys.NewService(cfg, st, newAPISync, common.NewDelayer(false)), st
}

func newKeysServiceWithNewAPI(t *testing.T) (domainkeys.Service, store.Store, *mock.StubAdminClient) {
	t.Helper()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, Key: "sk-test-key", RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	newAPISync := newapisync.New(cfg, st, newapi.NewAdminPortAdapter(stub), nil, newapisync.NewChannelPolicy(cfg))
	return domainkeys.NewService(cfg, st, newAPISync, common.NewDelayer(false)), st, stub
}

func newNewAPISync(t *testing.T, stub *mock.StubAdminClient) (*newapisync.NewAPISync, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	return newapisync.New(cfg, st, newapi.NewAdminPortAdapter(stub), nil, newapisync.NewChannelPolicy(cfg)), st
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
