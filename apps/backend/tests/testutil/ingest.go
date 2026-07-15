package testutil

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

func NewIngestService(t *testing.T, cfg config.Config, st store.Store) *usage.IngestService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	budgetOps := app.NewUsageBudgetOps(nil, nil, logger)
	lotConsumer := app.NewUsageLotConsumer()
	return usage.NewIngestService(cfg, st, st.Logs(), logger, app.NewUsageIngestEnqueuer(jobs.NoopEnqueuer{}), nil, budgetOps, lotConsumer)
}
