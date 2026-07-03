package app

import (
	"context"
	"fmt"
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
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/store"
)

func wireOrg(cfg config.Config, i infra, logger *slog.Logger) domainorg.Service {
	factory := datasource.NewFactory(cfg)
	return domainorg.NewService(cfg, i.store, factory, i.lifecycle, i.notifier, i.delayer, logger)
}

func wireBudget(cfg config.Config, i infra) domainbudget.Service {
	return domainbudget.NewService(cfg, i.store, i.delayer)
}

func wireOverrunService(cfg config.Config, i infra, logger *slog.Logger) domainbudget.OverrunProcessor {
	return domainbudget.NewOverrunService(cfg, i.store, i.lifecycle, i.notifier, logger)
}

func wireRebalance(cfg config.Config, i infra) domainbudget.Rebalancer {
	return domainbudget.NewRebalanceService(cfg, i.store, i.adminClient)
}

func wireKeys(cfg config.Config, i infra) domainkeys.Service {
	return domainkeys.NewService(cfg, i.store, i.lifecycle, i.delayer)
}

func wireModels(cfg config.Config, i infra) domainmodels.Service {
	return domainmodels.NewService(cfg, i.store, i.adminClient, i.lifecycle, i.delayer)
}

func wireDashboard(cfg config.Config, i infra, reader domainusage.Reader) domaindashboard.Service {
	return domaindashboard.NewService(cfg, i.store, reader)
}

func wireAudit(cfg config.Config, i infra) domainaudit.Service {
	return domainaudit.NewService(cfg, i.store)
}

func wireCompany(cfg config.Config, i infra) domaincompany.Service {
	return domaincompany.NewService(cfg, i.store, i.adminClient)
}

func wireBilling(cfg config.Config, i infra) domainbilling.Service {
	rebalanceEnqueue := func(ctx context.Context, companyID int64) error {
		return i.store.Relay().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	}
	return domainbilling.NewService(cfg, i.store, i.adminClient, i.wallet, rebalanceEnqueue)
}

func wireIngestService(cfg config.Config, i infra, logger *slog.Logger) *domainusage.IngestService {
	return domainusage.NewIngestService(cfg, i.store, i.notifier, logger)
}

func wireReader(i infra) domainusage.Reader {
	return domainusage.NewReader(i.store.Usage(), i.store.Ledger())
}
