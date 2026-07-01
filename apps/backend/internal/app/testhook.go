package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

func NewWithStore(cfg config.Config, logger *slog.Logger, st store.Store, opts ...Option) (*App, error) {
	return newApp(cfg, logger, st, opts...)
}
