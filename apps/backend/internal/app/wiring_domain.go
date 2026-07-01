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
	"github.com/tokenjoy/backend/internal/domain/session"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/store"
)

type domainServices struct {
	session   session.Service
	org       domainorg.Service
	budget    domainbudget.Service
	keys      domainkeys.Service
	models    domainmodels.Service
	dashboard domaindashboard.Service
	audit     domainaudit.Service
	ingest    domainbudget.Ingestor
	rebalance domainbudget.Rebalancer
	company   domaincompany.Service
	billing   domainbilling.Service
}

func buildDomainServices(cfg config.Config, i infra, logger *slog.Logger) domainServices {
	factory := datasource.NewFactory(cfg)
	logAggregator := domainusage.NewLogAggregator(i.adminClient, i.store, logger)
	ingest := domainbudget.NewIngestService(cfg, i.store, i.lifecycle, i.notifier, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, i.store, i.adminClient, i.lifecycle)
	companySvc := domaincompany.NewService(cfg, i.store, i.adminClient)
	rebalanceEnqueue := func(ctx context.Context, companyID int64) error {
		return i.store.Relay().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	}
	billingSvc := domainbilling.NewService(cfg, i.store, i.adminClient, i.wallet, rebalanceEnqueue)
	return domainServices{
		session:   session.NewService(i.store),
		org:       domainorg.NewService(cfg, i.store, factory, i.lifecycle, i.notifier, i.delayer, logger),
		budget:    domainbudget.NewService(cfg, i.store, i.delayer),
		keys:      domainkeys.NewService(cfg, i.store, i.lifecycle, i.delayer),
		models:    domainmodels.NewService(cfg, i.store, i.adminClient, i.lifecycle, i.delayer),
		dashboard: domaindashboard.NewService(cfg, i.store, logAggregator),
		audit:     domainaudit.NewService(cfg, i.store),
		ingest:    ingest,
		rebalance: rebalance,
		company:   companySvc,
		billing:   billingSvc,
	}
}
