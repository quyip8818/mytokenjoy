package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
)

func wireBudget(cfg config.Config, i infra) domainbudget.Service {
	return domainbudget.NewService(cfg, i.store, i.delayer)
}

func wireIngest(cfg config.Config, i infra, logger *slog.Logger) domainbudget.Ingestor {
	return domainbudget.NewIngestService(cfg, i.store, i.lifecycle, i.notifier, logger)
}

func wireRebalance(cfg config.Config, i infra) domainbudget.Rebalancer {
	return domainbudget.NewRebalanceService(cfg, i.store, i.adminClient, i.lifecycle)
}
