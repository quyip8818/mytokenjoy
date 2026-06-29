package testutil

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/store"
)

func NewIngestService(t *testing.T, cfg config.Config, st store.Store) *budget.IngestService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	notifier := notification.NewService(cfg, st, logger)
	return budget.NewIngestService(cfg, st, lifecycle, notifier, logger)
}
