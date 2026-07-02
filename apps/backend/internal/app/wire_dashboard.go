package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func wireDashboard(cfg config.Config, i infra, logger *slog.Logger) domaindashboard.Service {
	logAggregator := domainusage.NewLogAggregator(i.adminClient, i.store, logger)
	return domaindashboard.NewService(cfg, i.store, logAggregator)
}

func wireAudit(cfg config.Config, i infra) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store)
}
