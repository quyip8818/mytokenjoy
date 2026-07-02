package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

func wireOrg(cfg config.Config, i infra, logger *slog.Logger) domainorg.Service {
	factory := datasource.NewFactory(cfg)
	return domainorg.NewService(cfg, i.store, factory, i.lifecycle, i.notifier, i.delayer, logger)
}
