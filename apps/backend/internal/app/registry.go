package app

import (
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/devapi"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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

func ingestMetricsRecorder(cfg config.Config) ingestmetrics.Recorder {
	if cfg.IngestEnabled() {
		return ingestmetrics.NewCollector()
	}
	return ingestmetrics.NoopCollector()
}

func buildServiceRegistry(cfg config.Config, i infra, services domainServices, logger *slog.Logger, holder *jobs.Holder) ServiceRegistry {
	var gateway domaingateway.GatewayService
	if cfg.GatewayEnabled && cfg.NewAPIEnabled {
		gw, err := wireGatewayService(cfg, i, logger)
		if err != nil {
			panic(fmt.Errorf("wire gateway service: %w", err))
		}
		gateway = gw
	}
	authzSvc, credSvc, memberToken, platformToken, err := wireIdentity(cfg, i.store)
	if err != nil {
		panic(err)
	}
	metrics := ingestMetricsRecorder(cfg)
	var devBearer devapi.BearerResolver
	var devReadiness devapi.ReadinessChecker
	if sync, ok := i.newAPISync.(*newapisync.NewAPISync); ok {
		devBearer = sync
		devReadiness = sync
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
			IngestEnqueuer:       holder,
			IngestMetrics:        metrics,
			CompanySvc:           services.company,
			BillingSvc:           services.billing,
			MemberAnalyticsSvc:   services.memberAnalytics,
			WalletSvc:            i.wallet,
			CompanyGate:          i.companyGate,
			Gateway:              gateway,
			DevBearerResolver:    devBearer,
			DevReadinessChecker:  devReadiness,
			NotificationSvc:      i.notificationSvc,
			RateLimiter:          i.rateLimiter,
		},
		Infra:     i,
		OrgSync:   services.org,
		Overrun:   services.overrun,
		Rebalance: services.rebalance,
	}
}
