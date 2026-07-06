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
	domainmember "github.com/tokenjoy/backend/internal/domain/member"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

type domainServices struct {
	org       domainorg.Service
	budget    domainbudget.Service
	keys      domainkeys.Service
	models    domainmodels.Service
	dashboard domaindashboard.Service
	audit     domainaudit.Service
	readModel       domainusage.ReadModel
	ingest          domainusage.Ingestor
	failureRecorder domainusage.FailureRecorder
	overrun         domainbudget.OverrunProcessor
	rebalance domainbudget.Rebalancer
	company   domaincompany.Service
	billing   domainbilling.Service
	member    domainmember.Service
}

func wireFailureRecorder(i infra, logger *slog.Logger) domainusage.FailureRecorder {
	return domainusage.NewFailureRecorder(i.store.Logs(), logger)
}

func buildDomainServices(cfg config.Config, i infra, logger *slog.Logger) domainServices {
	reader := wireReader(i)
	failureRecorder := wireFailureRecorder(i, logger)
	keysSvc := wireKeys(cfg, i)
	return domainServices{
		org:             wireOrg(cfg, i, logger),
		budget:          wireBudget(cfg, i),
		keys:            keysSvc,
		models:          wireModels(cfg, i),
		dashboard:       wireDashboard(cfg, i, reader),
		audit:           wireAudit(cfg, i),
		readModel:       reader,
		ingest:          wireIngestService(cfg, i, logger),
		failureRecorder: failureRecorder,
		overrun:         wireOverrunService(cfg, i, logger),
		rebalance:       wireRebalance(cfg, i),
		company:         wireCompany(cfg, i),
		billing:         wireBilling(cfg, i),
		member:          wireMember(cfg, reader, keysSvc),
	}
}
