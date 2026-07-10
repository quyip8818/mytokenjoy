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
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

func wireOrg(cfg config.Config, i infra, logger *slog.Logger) domainorg.Service {
	factory := datasource.NewFactory(cfg)
	return domainorg.NewService(cfg, i.store, factory, i.newAPISync, i.notifier, i.delayer, logger)
}

func wireBudget(cfg config.Config, i infra) domainbudget.Service {
	return domainbudget.NewService(cfg, i.store, i.delayer)
}

func wireOverrunService(cfg config.Config, i infra, logger *slog.Logger) domainbudget.OverrunProcessor {
	return domainbudget.NewOverrunService(cfg, i.store, i.newAPISync, i.notifier, logger)
}

func wireRebalance(cfg config.Config, i infra) domainbudget.Rebalancer {
	return domainbudget.NewRebalanceService(cfg, i.store, i.adminClient)
}

func wireKeys(cfg config.Config, i infra) domainkeys.Service {
	return domainkeys.NewService(cfg, i.store, i.newAPISync, i.delayer)
}

func wireModels(cfg config.Config, i infra) domainmodels.Service {
	return domainmodels.NewService(cfg, i.store, i.adminClient, i.newAPISync, i.delayer)
}

func wireDashboard(cfg config.Config, i infra, reader domainusage.Reader) domaindashboard.Service {
	return domaindashboard.NewService(cfg, i.store, reader)
}

func wireAudit(cfg config.Config, i infra, reader domainusage.Reader) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store, reader)
}

func wireCompany(cfg config.Config, i infra) domaincompany.Service {
	return domaincompany.NewService(cfg, i.store, i.adminClient)
}

func wireBilling(cfg config.Config, i infra, reader domainusage.Reader) domainbilling.Service {
	return domainbilling.NewService(cfg, i.store, reader, i.adminClient, i.wallet, EnqueueRebalanceCompany(i.store), EnqueueWalletSync(i.store))
}

func wireMemberAnalytics(cfg config.Config, reader domainusage.Reader, keys domainkeys.Service) domainmemberanalytics.Service {
	return domainmemberanalytics.NewService(cfg, keys, reader)
}

func wireIngestService(cfg config.Config, i infra, logger *slog.Logger) *domainusage.IngestService {
	return domainusage.NewIngestService(cfg, i.store, i.store.Logs(), i.notifier, logger, EnqueueWalletSync(i.store))
}

func wireReader(i infra) domainusage.Reader {
	return domainusage.NewReader(i.store.Usage(), i.store.Ledger())
}
