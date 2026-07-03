package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/worker"
)

type ServiceRegistry struct {
	httpdeps.Deps
	Infra     infra
	OrgSync   domainorg.SyncService
	Overrun   domainbudget.OverrunProcessor
	Rebalance domainbudget.Rebalancer
}

func (r ServiceRegistry) HTTPDeps(logger *slog.Logger) httpdeps.Deps {
	d := r.Deps
	d.Logger = logger
	return d
}

func (r ServiceRegistry) WorkerRunner(logger *slog.Logger) *worker.Runner {
	return worker.NewRunner(
		r.Config,
		r.Store.Relay(),
		r.Infra.adminClient,
		r.Infra.lifecycle,
		r.IngestSvc,
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
	authzSvc, credSvc, memberToken, platformToken, err := wireIdentity(cfg, i.store)
	if err != nil {
		panic(err)
	}
	return ServiceRegistry{
		Deps: httpdeps.Deps{
			Config:               cfg,
			Store:                i.store,
			AuthzSvc:             authzSvc,
			Credentials:          credSvc,
			SessionToken:         memberToken,
			PlatformSessionToken: platformToken,
			OrgSvc:               services.org,
			BudgetSvc:            services.budget,
			KeysSvc:              services.keys,
			ModelsSvc:            services.models,
			DashboardSvc:         services.dashboard,
			AuditSvc:             services.audit,
			ReadModel:            services.readModel,
			IngestSvc:            services.ingest,
			CompanySvc:           services.company,
			BillingSvc:           services.billing,
			WalletSvc:            i.wallet,
			CompanyGate:          i.companyGate,
			RelayGateway:         relayGateway,
		},
		Infra:     i,
		OrgSync:   services.org,
		Overrun:   services.overrun,
		Rebalance: services.rebalance,
	}
}
