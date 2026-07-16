package testutil

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

func NewIngestService(t *testing.T, cfg config.Config, st store.Store) *usage.IngestService {
	t.Helper()
	return NewIngestServiceWithEnqueuer(t, cfg, st, jobs.NoopEnqueuer{})
}

// NewIngestServiceWithEnqueuer creates an IngestService with a custom enqueuer.
// Use this when testing enqueue failure or side-effect behavior.
func NewIngestServiceWithEnqueuer(t *testing.T, cfg config.Config, st store.Store, enqueuer jobs.Enqueuer) *usage.IngestService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	budgetOps := adapter.NewUsageBudgetOps(nil, nil, logger)
	lotConsumer := adapter.NewUsageLotConsumer()
	return usage.NewIngestService(cfg, st, st.Logs(), logger, adapter.NewUsageIngestEnqueuer(enqueuer), nil, budgetOps, lotConsumer)
}
