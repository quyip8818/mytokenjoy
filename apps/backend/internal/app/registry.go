package app

import (
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
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
		r.Store.AsyncJobs(),
		r.Store.SchedulerLock(),
		r.Store.Logs(),
		r.IngestMetrics,
		r.Infra.newAPISync,
		r.IngestSvc,
		r.IngestQueue,
		r.Overrun,
		r.Rebalance,
		r.BillingSvc,
		r.OrgSync,
		logger,
	)
}

func ingestMetricsRecorder(cfg config.Config) ingestmetrics.Recorder {
	if cfg.IngestEnabled() {
		return ingestmetrics.NewCollector()
	}
	return ingestmetrics.NoopCollector()
}

func buildServiceRegistry(cfg config.Config, i infra, services domainServices) ServiceRegistry {
	var newAPIGateway domaingateway.GatewayService
	if cfg.NewAPIGatewayEnabled && cfg.NewAPIEnabled {
		gw, err := wireGatewayService(cfg, i)
		if err != nil {
			panic(fmt.Errorf("wire gateway service: %w", err))
		}
		newAPIGateway = gw
	}
	authzSvc, credSvc, memberToken, platformToken, err := wireIdentity(cfg, i.store)
	if err != nil {
		panic(err)
	}
	metrics := ingestMetricsRecorder(cfg)
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
			IngestQueue:          services.ingestQueue,
			IngestMetrics:        metrics,
			CompanySvc:           services.company,
			BillingSvc:           services.billing,
			MemberAnalyticsSvc:   services.memberAnalytics,
			WalletSvc:            i.wallet,
			CompanyGate:          i.companyGate,
			NewAPIGateway:        newAPIGateway,
		},
		Infra:     i,
		OrgSync:   services.org,
		Overrun:   services.overrun,
		Rebalance: services.rebalance,
	}
}
