package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newIngestStore(t *testing.T) (config.Config, store.Store) {
	t.Helper()
	return testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
}
