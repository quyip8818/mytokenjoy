package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domaingrants "github.com/tokenjoy/backend/internal/domain/grants"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

func dashboardScopeConfig() domainusage.DashboardScopeConfig {
	return domainusage.DashboardScopeConfig{
		OrgWidePermissions: []string{permission.DashboardCost, permission.DashboardUsage},
	}
}

func wireOrg(cfg config.Config, i infra, logger *slog.Logger, grants domaingrants.Normalizer, enqueuer jobs.Enqueuer, orgAdmin *adapter.OrgRiverAdminHolder) domainorg.Service {
	factory := datasource.NewFactory(cfg)
	return domainorg.NewService(cfg, i.store, factory, i.newAPISync, i.notifier, i.delayer, logger, grants, adapter.NewOrgEnqueuer(enqueuer, orgAdmin))
}

func wireBudget(cfg config.Config, i infra, enqueuer jobs.Enqueuer) domainbudget.Service {
	return domainbudget.NewService(cfg, i.store, i.delayer, adapter.NewBudgetEnqueuer(enqueuer))
}

func wireOverrunService(cfg config.Config, i infra, logger *slog.Logger) domainbudget.OverrunProcessor {
	return domainbudget.NewOverrunService(cfg, i.store, i.newAPISync, i.notifier, logger)
}

func wireRebalance(cfg config.Config, i infra) domainbudget.Rebalancer {
	cache := budgetcheck.WrapStore(i.budgetCheck)
	return domainbudget.NewRebalanceService(cfg, i.store, domainbudget.WithRebalanceCache(cache))
}

func wireKeys(cfg config.Config, i infra) domainkeys.Service {
	return domainkeys.NewService(cfg, i.store, i.newAPISync, i.delayer, domainkeys.WithCacheInvalidator(i.precheckCache))
}

func wireModels(cfg config.Config, i infra) domainmodels.Service {
	return domainmodels.NewService(cfg, i.store, i.adminPort, i.newAPISync, i.delayer)
}

func wireDashboard(cfg config.Config, i infra, reader domainusage.Reader) domaindashboard.Service {
	return domaindashboard.NewService(cfg, i.store, reader, dashboardScopeConfig())
}

func wireAudit(cfg config.Config, i infra, reader domainusage.Reader) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store, reader)
}

func wireCompany(cfg config.Config, i infra, grants domaingrants.Normalizer) domaincompany.Service {
	return domaincompany.NewService(cfg, i.store, i.adminPort, grants, domaincompany.WithCompanyCacheInvalidator(i.precheckCache))
}

func wireBilling(cfg config.Config, i infra, reader domainusage.Reader) domainbilling.Service {
	return domainbilling.NewService(cfg, i.store, reader)
}

func wireMemberAnalytics(cfg config.Config, reader domainusage.Reader, keys domainkeys.Service) domainmemberanalytics.Service {
	return domainmemberanalytics.NewService(cfg, keys, reader)
}

func wireIngestService(cfg config.Config, i infra, logger *slog.Logger, enqueuer jobs.Enqueuer) *domainusage.IngestService {
	alertPub := adapter.NewBudgetAlertPublisher(i.notificationSvc)
	cache := budgetcheck.WrapStore(i.budgetCheck)
	budgetOps := adapter.NewUsageBudgetOps(cache, alertPub, logger)
	lotConsumer := adapter.NewUsageLotConsumer()
	return domainusage.NewIngestService(cfg, i.store, i.store.Logs(), logger, adapter.NewUsageIngestEnqueuer(enqueuer), i.notifier, budgetOps, lotConsumer)
}

func wireReader(i infra) domainusage.Reader {
	return domainusage.NewReader(i.store.Usage(), i.store.Ledger())
}
