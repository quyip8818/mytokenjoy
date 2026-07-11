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
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type domainServices struct {
	org             domainorg.Service
	budget          domainbudget.Service
	keys            domainkeys.Service
	models          domainmodels.Service
	dashboard       domaindashboard.Service
	audit           domainaudit.Service
	readModel       domainusage.ReadModel
	ingest          domainusage.Ingestor
	ingestQueue     domainusage.Queue
	overrun         domainbudget.OverrunProcessor
	rebalance       domainbudget.Rebalancer
	company         domaincompany.Service
	billing         domainbilling.Service
	memberAnalytics domainmemberanalytics.Service
}

func wireIngestQueue(i infra) domainusage.Queue {
	return domainusage.NewQueue(i.store.Logs())
}

func buildDomainServices(cfg config.Config, i infra, logger *slog.Logger) domainServices {
	reader := wireReader(i)
	ingestQueue := wireIngestQueue(i)
	keysSvc := wireKeys(cfg, i)
	grants := permission.NewGrantNormalizer()
	return domainServices{
		org:             wireOrg(cfg, i, logger, grants),
		budget:          wireBudget(cfg, i),
		keys:            keysSvc,
		models:          wireModels(cfg, i),
		dashboard:       wireDashboard(cfg, i, reader),
		audit:           wireAudit(cfg, i, reader),
		readModel:       reader,
		ingest:          wireIngestService(cfg, i, logger),
		ingestQueue:     ingestQueue,
		overrun:         wireOverrunService(cfg, i, logger),
		rebalance:       wireRebalance(cfg, i),
		company:         wireCompany(cfg, i, grants),
		billing:         wireBilling(cfg, i, reader),
		memberAnalytics: wireMemberAnalytics(cfg, reader, keysSvc),
	}
}
