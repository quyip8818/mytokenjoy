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
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/session"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/platformauth"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
)

type ServiceRegistry struct {
	Config         config.Config
	Store          store.Store
	Infra          infra
	Session        session.Service
	Org            domainorg.Service
	OrgSync        domainorg.SyncService
	Budget         domainbudget.Service
	Keys           domainkeys.Service
	Models         domainmodels.Service
	Dashboard      domaindashboard.Service
	Audit          domainaudit.Service
	ReadModel      domainusage.ReadModel
	Ingest         domainusage.Ingestor
	Overrun        domainbudget.OverrunProcessor
	Rebalance      domainbudget.Rebalancer
	Company        domaincompany.Service
	Billing        domainbilling.Service
	Platform       platformauth.Service
	CompanyGate    *domaincompany.Gate
	RelayGateway   domainrelay.GatewayService
}

func (r ServiceRegistry) HTTPDeps(logger *slog.Logger) httpdeps.Deps {
	return httpdeps.Deps{
		Config:       r.Config,
		Logger:       logger,
		SessionSvc:   r.Session,
		OrgSvc:       r.Org,
		BudgetSvc:    r.Budget,
		KeysSvc:      r.Keys,
		ModelsSvc:    r.Models,
		DashboardSvc: r.Dashboard,
		AuditSvc:     r.Audit,
		ReadModel:    r.ReadModel,
		IngestSvc:    r.Ingest,
		CompanySvc:   r.Company,
		BillingSvc:   r.Billing,
		PlatformSvc:  r.Platform,
		WalletSvc:    r.Infra.wallet,
		CompanyGate:  r.CompanyGate,
		RelayGateway: r.RelayGateway,
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
		r.OrgSync,
		logger,
	)
}

func buildServiceRegistry(cfg config.Config, i infra, services domainServices) ServiceRegistry {
	var relayGateway domainrelay.GatewayService
	if cfg.RelayGatewayEnabled && cfg.NewAPIEnabled {
		gw, err := wireGatewayService(cfg, i)
		if err == nil {
			relayGateway = gw
		}
	}
	return ServiceRegistry{
		Config:       cfg,
		Store:        i.store,
		Infra:        i,
		Session:      services.session,
		Org:          services.org,
		OrgSync:      services.org,
		Budget:       services.budget,
		Keys:         services.keys,
		Models:       services.models,
		Dashboard:    services.dashboard,
		Audit:        services.audit,
		ReadModel:    services.readModel,
		Ingest:       services.ingest,
		Overrun:      services.overrun,
		Rebalance:    services.rebalance,
		Company:      services.company,
		Billing:      services.billing,
		Platform:     platformauth.NewService(cfg, i.store),
		CompanyGate:  i.companyGate,
		RelayGateway: relayGateway,
	}
}
