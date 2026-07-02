package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func wireDashboard(cfg config.Config, i infra) domaindashboard.Service {
	return domaindashboard.NewService(cfg, i.store)
}

func wireAudit(cfg config.Config, i infra) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store)
}

func wireCallLogQuerier(i infra) domainusage.CallLogQuerier {
	return domainusage.NewCallLogQuerier(i.store)
}
