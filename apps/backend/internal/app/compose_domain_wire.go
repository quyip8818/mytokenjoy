package app

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/adapter/bridge"
	"github.com/tokenjoy/backend/internal/adapter/enqueue"
	"github.com/tokenjoy/backend/internal/config"
	domainapproval "github.com/tokenjoy/backend/internal/domain/approval"
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
	"github.com/tokenjoy/backend/internal/pkg/invitetoken"
	"github.com/tokenjoy/backend/internal/store"
)

func dashboardScopeConfig() domainusage.DashboardScopeConfig {
	return domainusage.DashboardScopeConfig{
		OrgWidePermissions: []string{permission.DashboardCost, permission.DashboardUsage},
	}
}

func wireOrg(cfg config.Config, i infra, logger *slog.Logger, grants domaingrants.Normalizer, enqueuer jobs.Enqueuer, orgAdmin *enqueue.OrgRiverAdminHolder) domainorg.Service {
	factory := datasource.NewFactory(cfg)
	return domainorg.NewService(cfg, i.store, factory, i.notifier, i.notificationSvc, i.delayer, logger, grants, enqueue.NewOrgEnqueuer(enqueuer, orgAdmin))
}

func wireBudget(cfg config.Config, i infra, enqueuer jobs.Enqueuer) domainbudget.Service {
	return domainbudget.NewService(cfg, i.store, i.delayer, enqueue.NewBudgetEnqueuer(enqueuer))
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
	return domainmodels.NewService(cfg, i.store, i.adminPort, i.precheckCache, i.delayer)
}

func wireDashboard(cfg config.Config, i infra, reader domainusage.Reader) domaindashboard.Service {
	return domaindashboard.NewService(cfg, i.store, reader, dashboardScopeConfig())
}

func wireAudit(cfg config.Config, i infra, reader domainusage.Reader) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store, reader)
}

func wireCompany(cfg config.Config, i infra, grants domaingrants.Normalizer) domaincompany.Service {
	opts := []domaincompany.CompanyServiceOption{
		domaincompany.WithCompanyCacheInvalidator(i.precheckCache),
		domaincompany.WithEmailSender(i.notificationSvc),
	}
	if keys := cfg.InviteSecretKeys(); len(keys) > 0 {
		if iss, err := invitetoken.NewIssuer(keys...); err == nil {
			opts = append(opts, domaincompany.WithInviteIssuer(iss))
		}
	}
	return domaincompany.NewService(cfg, i.store, i.adminPort, grants, opts...)
}

func wireBilling(cfg config.Config, i infra, reader domainusage.Reader) domainbilling.Service {
	return domainbilling.NewService(cfg, i.store, reader, i.adminPort)
}

func wireMemberAnalytics(cfg config.Config, reader domainusage.Reader, budget domainbudget.Service) domainmemberanalytics.Service {
	return domainmemberanalytics.NewService(cfg, budget, reader)
}

func wireIngestService(cfg config.Config, i infra, logger *slog.Logger, enqueuer jobs.Enqueuer) *domainusage.IngestService {
	alertPub := bridge.NewBudgetAlertPublisher(i.notificationSvc)
	cache := budgetcheck.WrapStore(i.budgetCheck)
	budgetOps := bridge.NewUsageBudgetOps(cache, alertPub, logger)
	lotConsumer := bridge.NewUsageLotConsumer()
	return domainusage.NewIngestService(cfg, i.store, i.store.Logs(), logger, enqueue.NewUsageIngestEnqueuer(enqueuer), i.notifier, budgetOps, lotConsumer)
}

func wireReader(i infra) domainusage.Reader {
	return domainusage.NewReader(i.store.Usage(), i.store.Ledger())
}

func wireApprovalEngine(i infra, logger *slog.Logger, keysSvc domainkeys.Service, budgetSvc domainbudget.Service) *domainapproval.Engine {
	repo := i.store.Approval()
	txRunner := func(ctx context.Context, fn func(store.Store) error) error {
		return i.store.WithTx(ctx, fn)
	}
	return domainapproval.NewEngine(repo, txRunner, logger,
		domainkeys.NewKeyApprovalHandler(keysSvc),
		domainbudget.NewMemberBudgetApprovalHandler(budgetSvc),
		domainbudget.NewProjectBudgetApprovalHandler(budgetSvc),
		domainbudget.NewProjectMemberBudgetApprovalHandler(budgetSvc),
	)
}
