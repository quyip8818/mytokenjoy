package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
)

type ServiceRegistry struct {
	Config    config.Config
	Store     store.Store
	Infra     infra
	Session   session.Service
	Org       domainorg.Service
	Budget    domainbudget.Service
	Keys      domainkeys.Service
	Models    domainmodels.Service
	Dashboard domaindashboard.Service
	Audit     domainaudit.Service
	Ingest    domainbudget.Ingestor
	Rebalance domainbudget.Rebalancer
}

func (r ServiceRegistry) HTTPDeps(logger *slog.Logger) httpapi.Deps {
	return httpapi.Deps{
		Config:       r.Config,
		Logger:       logger,
		SessionSvc:   r.Session,
		OrgSvc:       r.Org,
		BudgetSvc:    r.Budget,
		KeysSvc:      r.Keys,
		ModelsSvc:    r.Models,
		DashboardSvc: r.Dashboard,
		AuditSvc:     r.Audit,
		IngestSvc:    r.Ingest,
	}
}

func (r ServiceRegistry) WorkerRunner(logger *slog.Logger) *worker.Runner {
	return worker.NewRunner(
		r.Config,
		r.Store,
		r.Infra.adminClient,
		r.Infra.lifecycle,
		r.Ingest,
		r.Rebalance,
		r.Org,
		logger,
	)
}

func buildServiceRegistry(cfg config.Config, i infra, services domainServices) ServiceRegistry {
	return ServiceRegistry{
		Config:    cfg,
		Store:     i.store,
		Infra:     i,
		Session:   services.session,
		Org:       services.org,
		Budget:    services.budget,
		Keys:      services.keys,
		Models:    services.models,
		Dashboard: services.dashboard,
		Audit:     services.audit,
		Ingest:    services.ingest,
		Rebalance: services.rebalance,
	}
}
