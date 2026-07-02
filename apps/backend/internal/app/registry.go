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
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/platformauth"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
)

type ServiceRegistry struct {
	Config      config.Config
	Store       store.Store
	Infra       infra
	Session     session.Service
	Org         domainorg.Service
	Budget      domainbudget.Service
	Keys        domainkeys.Service
	Models      domainmodels.Service
	Dashboard   domaindashboard.Service
	Audit           domainaudit.Service
	CallLogQuerier  domainusage.CallLogQuerier
	Ingest          domainbudget.Ingestor
	Overrun     domainbudget.OverrunProcessor
	Rebalance   domainbudget.Rebalancer
	Company     domaincompany.Service
	Billing     domainbilling.Service
	Platform    platformauth.Service
	CompanyGate *domaincompany.Gate
}

func (r ServiceRegistry) HTTPDeps(logger *slog.Logger) httpdeps.Deps {
	return httpdeps.Deps{
		Config:       r.Config,
		Logger:       logger,
		Store:        r.Store,
		SessionSvc:   r.Session,
		OrgSvc:       r.Org,
		BudgetSvc:    r.Budget,
		KeysSvc:      r.Keys,
		ModelsSvc:    r.Models,
		DashboardSvc: r.Dashboard,
		AuditSvc:         r.Audit,
		CallLogQuerier:   r.CallLogQuerier,
		IngestSvc:        r.Ingest,
		CompanySvc:   r.Company,
		BillingSvc:   r.Billing,
		PlatformSvc:  r.Platform,
		WalletSvc:    r.Infra.wallet,
		CompanyGate:  r.CompanyGate,
	}
}

func (r ServiceRegistry) WorkerRunner(logger *slog.Logger) *worker.Runner {
	return worker.NewRunner(
		r.Config,
		r.Store,
		r.Infra.adminClient,
		r.Infra.lifecycle,
		r.Ingest,
		r.Overrun,
		r.Rebalance,
		r.Org,
		logger,
	)
}

func buildServiceRegistry(cfg config.Config, i infra, services domainServices) ServiceRegistry {
	return ServiceRegistry{
		Config:      cfg,
		Store:       i.store,
		Infra:       i,
		Session:     services.session,
		Org:         services.org,
		Budget:      services.budget,
		Keys:        services.keys,
		Models:      services.models,
		Dashboard:   services.dashboard,
		Audit:           services.audit,
		CallLogQuerier:  services.callLogQuerier,
		Ingest:          services.ingest,
		Overrun:     services.overrun,
		Rebalance:   services.rebalance,
		Company:     services.company,
		Billing:     services.billing,
		Platform:    platformauth.NewService(cfg, i.store),
		CompanyGate: i.companyGate,
	}
}
