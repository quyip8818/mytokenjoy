package worker

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

type Runner struct {
	cfg            config.Config
	relayOutbox    store.RelayOutboxRepository
	webhookOutbox  store.WebhookOutboxRepository
	rebalanceQueue store.RebalanceQueueRepository
	overrunQueue   store.OverrunQueueRepository
	syncCursor     store.SyncCursorRepository
	lifecycle      relay.Lifecycle
	ingest         domainbudget.Ingestor
	overrun        domainbudget.OverrunProcessor
	rebalance      domainbudget.Rebalancer
	syncSvc        domainorg.SyncService
	client         newapi.AdminClient
	logger         *slog.Logger
	interval       time.Duration
	syncEvery      time.Duration
	syncTick       time.Duration
}

func NewRunner(
	cfg config.Config,
	st store.Store,
	client newapi.AdminClient,
	lifecycle relay.Lifecycle,
	ingest domainbudget.Ingestor,
	overrun domainbudget.OverrunProcessor,
	rebalance domainbudget.Rebalancer,
	syncSvc domainorg.SyncService,
	logger *slog.Logger,
) *Runner {
	relayRepo := st.Relay()
	return &Runner{
		cfg:            cfg,
		relayOutbox:    relayRepo,
		webhookOutbox:  relayRepo,
		rebalanceQueue: relayRepo,
		overrunQueue:   relayRepo,
		syncCursor:     relayRepo,
		client:         client,
		lifecycle:      lifecycle,
		ingest:         ingest,
		overrun:        overrun,
		rebalance:      rebalance,
		syncSvc:        syncSvc,
		logger:         logger,
		interval:       cfg.WorkerPollInterval(),
		syncEvery:      cfg.WorkerOrgSyncInterval(),
	}
}

func (r *Runner) Start(ctx context.Context) {
	if !r.cfg.NewAPIEnabled {
		return
	}
	go r.loop(ctx)
}

func (r *Runner) loop(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.syncTick += r.interval
			r.tick(ctx)
			if r.syncTick >= r.syncEvery {
				r.syncTick = 0
				r.logStep("org_sync", r.processOrgSync(ctx))
			}
		}
	}
}

func (r *Runner) tick(ctx context.Context) {
	r.logStep("relay_outbox", r.processRelayOutbox(ctx))
	r.logStep("webhook_outbox", r.processWebhookOutbox(ctx))
	r.logStep("rebalance", r.processRebalance(ctx))
	r.logStep("overrun", r.processOverrun(ctx))
	r.logStep("compensate_logs", r.compensateLogs(ctx))
}

func (r *Runner) logStep(step string, err error) {
	if err != nil {
		r.logger.Warn("worker step failed", "step", step, "error", err)
	}
}

// RunOnce executes one worker cycle (outbox + rebalance + log compensation).
func (r *Runner) RunOnce(ctx context.Context) { r.tick(ctx) }

func backoff(attempts int) time.Duration {
	seconds := math.Min(300, math.Pow(2, float64(attempts)))
	return time.Duration(seconds) * time.Second
}

func (r *Runner) markRelayOutboxRetry(ctx context.Context, id string, next time.Time, reason string) {
	if err := r.relayOutbox.MarkRelayOutboxRetry(ctx, id, next, reason); err != nil {
		r.logger.Warn("mark relay outbox retry failed", "id", id, "error", err)
	}
}

func (r *Runner) markRelayOutboxDone(ctx context.Context, id string) {
	if err := r.relayOutbox.MarkRelayOutboxDone(ctx, id); err != nil {
		r.logger.Warn("mark relay outbox done failed", "id", id, "error", err)
	}
}

func (r *Runner) markWebhookOutboxRetry(ctx context.Context, id string, next time.Time, reason string) {
	if err := r.webhookOutbox.MarkWebhookOutboxRetry(ctx, id, next, reason); err != nil {
		r.logger.Warn("mark webhook outbox retry failed", "id", id, "error", err)
	}
}

func (r *Runner) markWebhookOutboxDone(ctx context.Context, id string) {
	if err := r.webhookOutbox.MarkWebhookOutboxDone(ctx, id); err != nil {
		r.logger.Warn("mark webhook outbox done failed", "id", id, "error", err)
	}
}

func (r *Runner) markRebalanceDone(ctx context.Context, id string) {
	if err := r.rebalanceQueue.MarkRebalanceDone(ctx, id); err != nil {
		r.logger.Warn("mark rebalance done failed", "id", id, "error", err)
	}
}
