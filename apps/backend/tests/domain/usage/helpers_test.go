package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func newIngestStore(t *testing.T) (config.Config, store.Store) {
	t.Helper()
	return testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
}

// ingestFixture provides a ready-to-use IngestService with mapping and budget
// headroom pre-seeded. Use newIngestFixture for most domain-level ingest tests.
type ingestFixture struct {
	Cfg    config.Config
	Store  store.Store
	Ingest *usage.IngestService
}

type ingestFixtureOption func(*ingestFixtureConfig)

type ingestFixtureConfig struct {
	mappingOpts newapisynctf.MappingOpts
	amount      *int64
	skipMapping bool
	enqueuer    jobs.Enqueuer
}

// withMappingOpts overrides the default mapping options.
func withMappingOpts(opts newapisynctf.MappingOpts) ingestFixtureOption {
	return func(c *ingestFixtureConfig) {
		c.mappingOpts = opts
	}
}

// withBudgetAmount overrides the budget headroom amount.
func withBudgetAmount(amount int64) ingestFixtureOption {
	return func(c *ingestFixtureConfig) {
		c.amount = &amount
	}
}

// withoutMapping skips PrepareIngestFixture (for tests that verify missing-mapping errors).
func withoutMapping() ingestFixtureOption {
	return func(c *ingestFixtureConfig) {
		c.skipMapping = true
	}
}

// withEnqueuer overrides the default NoopEnqueuer (e.g. for testing enqueue failures).
func withEnqueuer(e jobs.Enqueuer) ingestFixtureOption {
	return func(c *ingestFixtureConfig) {
		c.enqueuer = e
	}
}

// newIngestFixture creates a pure-domain ingest fixture with PrepareIngestFixture
// pre-applied (unless withoutMapping is used).
func newIngestFixture(t *testing.T, opts ...ingestFixtureOption) ingestFixture {
	t.Helper()
	fc := ingestFixtureConfig{
		mappingOpts: newapisynctf.DefaultMappingOpts(),
	}
	for _, o := range opts {
		o(&fc)
	}
	cfg, st := newIngestStore(t)
	var ingest *usage.IngestService
	if fc.enqueuer != nil {
		ingest = testutil.NewIngestServiceWithEnqueuer(t, cfg, st, fc.enqueuer)
	} else {
		ingest = testutil.NewIngestService(t, cfg, st)
	}
	if !fc.skipMapping {
		if fc.amount != nil {
			newapisynctf.PrepareIngestFixture(t, st, fc.mappingOpts, *fc.amount)
		} else {
			newapisynctf.PrepareIngestFixture(t, st, fc.mappingOpts)
		}
	}
	return ingestFixture{Cfg: cfg, Store: st, Ingest: ingest}
}
