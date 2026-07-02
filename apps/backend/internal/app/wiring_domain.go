package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/session"
)

type domainServices struct {
	session       session.Service
	org           domainorg.Service
	budget        domainbudget.Service
	keys          domainkeys.Service
	models        domainmodels.Service
	dashboard     domaindashboard.Service
	audit         domainaudit.Service
	callLogQuerier domainusage.CallLogQuerier
	ingest        domainbudget.Ingestor
	overrun       domainbudget.OverrunProcessor
	rebalance     domainbudget.Rebalancer
	company       domaincompany.Service
	billing       domainbilling.Service
}

func buildDomainServices(cfg config.Config, i infra, logger *slog.Logger) domainServices {
	return domainServices{
		session:        wireSession(i.store),
		org:            wireOrg(cfg, i, logger),
		budget:         wireBudget(cfg, i),
		keys:           wireKeys(cfg, i),
		models:         wireModels(cfg, i),
		dashboard:      wireDashboard(cfg, i),
		audit:          wireAudit(cfg, i),
		callLogQuerier: wireCallLogQuerier(i),
		ingest:         wireIngestService(cfg, i, logger),
		overrun:        wireOverrunService(cfg, i, logger),
		rebalance:      wireRebalance(cfg, i),
		company:        wireCompany(cfg, i),
		billing:        wireBilling(cfg, i),
	}
}
