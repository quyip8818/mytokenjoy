//go:build testhook

package riverfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

// NewInsertOnlyEnqueuer returns an enqueuer backed by a River client that can insert jobs
// without starting workers (for domain tests that only assert enqueue side effects).
func NewInsertOnlyEnqueuer(t *testing.T, cfg config.Config, st store.Store) jobs.Enqueuer {
	t.Helper()
	pool := postgres.MainPool(st)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return client.Enqueuer
}
