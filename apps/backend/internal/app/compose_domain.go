package app

import (
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/adapter/enqueue"
	"github.com/tokenjoy/backend/internal/config"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/grants"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapisync"
	"github.com/tokenjoy/backend/internal/integration/newapisync/devapi"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

func ingestMetricsRecorder(cfg config.Config) ingestmetrics.Recorder {
	if cfg.IngestEnabled() {
		return ingestmetrics.NewCollector()
	}
	return ingestmetrics.NoopCollector()
}

// buildServiceRegistry constructs all domain services and wires them into the
// ServiceRegistry (which embeds httpdeps.Deps for the HTTP layer and holds
// worker-only fields separately).
func buildServiceRegistry(cfg config.Config, i infra, logger *slog.Logger, holder *jobs.Holder, orgAdmin *enqueue.OrgRiverAdminHolder, overrideClk clock.Clock) ServiceRegistry {
	clk := clock.OrDefault(overrideClk)

	// --- Domain services ---
	reader := wireReader(i)
	keysSvc := wireKeys(cfg, i, clk)
	budgetSvc := wireBudget(cfg, i, holder, clk)
	normalizer := grants.NewGrantNormalizer()
	orgSvc := wireOrg(cfg, i, logger, normalizer, holder, orgAdmin, clk)

	// --- Identity ---
	authzSvc, credSvc, memberToken, err := wireIdentity(cfg, i.store)
	if err != nil {
		panic(err)
	}

	// --- Gateway ---
	var gateway domaingateway.GatewayService
	if cfg.GatewayEnabled && cfg.NewAPIEnabled {
		gw, err := wireGatewayService(cfg, i, logger, clk)
		if err != nil {
			panic(fmt.Errorf("wire gateway service: %w", err))
		}
		gateway = gw
	}

	// --- Dev API (only if NewAPISync concrete type) ---
	var devBearer devapi.BearerResolver
	var devReadiness devapi.ReadinessChecker
	if sync, ok := i.newAPISync.(*newapisync.NewAPISync); ok {
		devBearer = sync
		devReadiness = sync
	}

	return ServiceRegistry{
		Deps: httpdeps.Deps{
			Config:              cfg,
			Logger:              logger,
			Store:               i.store,
			AdminPort:           i.adminPort,
			AuthzSvc:            authzSvc,
			Credentials:         credSvc,
			SessionToken:        memberToken,
			OrgSvc:              orgSvc,
			BudgetSvc:           budgetSvc,
			KeysSvc:             keysSvc,
			ModelsSvc:           wireModels(cfg, i),
			DashboardSvc:        wireDashboard(cfg, i, reader, clk),
			AuditSvc:            wireAudit(cfg, i, reader),
			ReadModel:           reader,
			IngestSvc:           wireIngestService(cfg, i, logger, holder, clk),
			IngestEnqueuer:      holder,
			IngestMetrics:       ingestMetricsRecorder(cfg),
			CompanySvc:          wireCompany(cfg, i, normalizer),
			BillingSvc:          wireBilling(i, reader),
			MemberAnalyticsSvc:  wireMemberAnalytics(cfg, reader, budgetSvc, clk),
			CompanyGate:         i.companyGate,
			ApprovalEngine:      wireApprovalEngine(i, logger, keysSvc, budgetSvc),
			Gateway:             gateway,
			DevBearerResolver:   devBearer,
			DevReadinessChecker: devReadiness,
			NotificationSvc:     i.notificationSvc,
			RateLimiter:         i.rateLimiter,
			VerifyCodeSvc:       i.verifyCodeSvc,
		},
		Infra:     i,
		OrgSync:   orgSvc,
		Overrun:   wireOverrunService(cfg, i, logger, clk),
		Rebalance: wireRebalance(cfg, i, clk),
	}
}
