package testutil

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/store"
)

func NewIngestService(t *testing.T, cfg config.Config, st store.Store) *usage.IngestService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	return usage.NewIngestService(cfg, st, notifier, logger)
}
