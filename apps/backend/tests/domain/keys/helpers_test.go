package keys_test

import (
	"testing"

	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newKeysService(t *testing.T) (domainkeys.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil, nil, relay.NewChannelPolicy(cfg))
	return domainkeys.NewService(cfg, st, lifecycle, common.NewDelayer(false)), st
}

func newKeysServiceWithRelay(t *testing.T) (domainkeys.Service, store.Store, *mock.StubAdminClient) {
	t.Helper()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, Key: "sk-test-key", RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	return domainkeys.NewService(cfg, st, lifecycle, common.NewDelayer(false)), st, stub
}

func newTokenLifecycle(t *testing.T, stub *mock.StubAdminClient) (*relay.TokenLifecycle, store.Store) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	return lifecycle, st
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
