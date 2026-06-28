package keys_test

import (
	"testing"

	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newKeysService(t *testing.T) (domainkeys.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	return domainkeys.NewService(cfg, st, lifecycle), st
}

func newTokenLifecycle(t *testing.T, stub *mock.StubAdminClient) (*relay.TokenLifecycle, store.Store) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub)
	return lifecycle, st
}

func findApproval(st store.Store, id string) *types.KeyApproval {
	for _, a := range st.Keys().Approvals() {
		if a.ID == id {
			copy := a
			return &copy
		}
	}
	return nil
}
