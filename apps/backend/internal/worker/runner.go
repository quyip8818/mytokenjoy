package worker

import (
	"context"
	"encoding/json"
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
	cfg       config.Config
	store     store.Store
	lifecycle relay.Lifecycle
	ingest    domainbudget.Ingestor
	rebalance domainbudget.Rebalancer
	syncSvc   domainorg.SyncService
	client    newapi.AdminClient
	logger    *slog.Logger
	interval  time.Duration
	syncEvery time.Duration
	syncTick  time.Duration
}

func NewRunner(
	cfg config.Config,
	st store.Store,
	client newapi.AdminClient,
	lifecycle relay.Lifecycle,
	ingest domainbudget.Ingestor,
	rebalance domainbudget.Rebalancer,
	syncSvc domainorg.SyncService,
	logger *slog.Logger,
) *Runner {
	return &Runner{
		cfg:       cfg,
		store:     st,
		client:    client,
		lifecycle: lifecycle,
		ingest:    ingest,
		rebalance: rebalance,
		syncSvc:   syncSvc,
		logger:    logger,
		interval:  5 * time.Second,
		syncEvery: time.Minute,
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
	r.logStep("compensate_logs", r.compensateLogs(ctx))
}

func (r *Runner) logStep(step string, err error) {
	if err != nil {
		r.logger.Warn("worker step failed", "step", step, "error", err)
	}
}

func (r *Runner) processOrgSync(ctx context.Context) error {
	if r.syncSvc == nil {
		return nil
	}
	return r.syncSvc.RunScheduledSync(ctx)
}

// RunOnce executes one worker cycle (outbox + rebalance + log compensation).
func (r *Runner) RunOnce(ctx context.Context) { r.tick(ctx) }

func (r *Runner) processRelayOutbox(ctx context.Context) error {
	entries, err := r.store.Relay().ClaimPendingRelayOutbox(20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		var processErr error
		switch entry.Kind {
		case store.OutboxKindCreateToken:
			var payload relay.CreateTokenOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			_, processErr = r.lifecycle.TrySyncCreate(ctx, payload.PlatformKeyID)
		case store.OutboxKindUpdateToken:
			var payload relay.UpdateTokenOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			processErr = r.lifecycle.SyncUpdatePlatformKey(ctx, payload.PlatformKeyID)
		case store.OutboxKindRevokeToken:
			var payload relay.UpdateTokenOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			processErr = r.lifecycle.SyncRevokePlatformKey(ctx, payload.PlatformKeyID)
		case store.OutboxKindUpsertChannel:
			var payload relay.UpsertChannelOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			processErr = r.lifecycle.SyncUpsertProviderKey(ctx, payload.ProviderKeyID)
		case store.OutboxKindUpdateModelLimits:
			var payload relay.UpdateModelLimitsOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			processErr = r.lifecycle.SyncModelLimitsForDepartment(ctx, payload.DepartmentID)
		case store.OutboxKindRebalanceToken:
			var payload relay.RebalanceAxisOutboxPayload
			_ = json.Unmarshal(entry.Payload, &payload)
			processErr = r.rebalance.ProcessAxis(ctx, payload.AxisKind, payload.AxisID)
		default:
			processErr = nil
		}
		if processErr != nil {
			next := time.Now().Add(backoff(entry.Attempts))
			_ = r.store.Relay().MarkRelayOutboxRetry(entry.ID, next, processErr.Error())
			continue
		}
		_ = r.store.Relay().MarkRelayOutboxDone(entry.ID)
	}
	return nil
}

func (r *Runner) processWebhookOutbox(ctx context.Context) error {
	entries, err := r.store.Relay().ClaimPendingWebhookOutbox(20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := r.ingest.IngestFromOutbox(ctx, entry.Payload); err != nil {
			next := time.Now().Add(backoff(entry.Attempts))
			_ = r.store.Relay().MarkWebhookOutboxRetry(entry.ID, next, err.Error())
			continue
		}
		_ = r.store.Relay().MarkWebhookOutboxDone(entry.ID)
	}
	return nil
}

func (r *Runner) processRebalance(ctx context.Context) error {
	entries, err := r.store.Relay().ClaimPendingRebalance(20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := r.rebalance.ProcessAxis(ctx, entry.AxisKind, entry.AxisID); err != nil {
			r.logger.Warn("rebalance failed", "axis", entry.AxisKind, "id", entry.AxisID, "error", err)
			_ = r.lifecycle.EnqueueRebalanceAxis(entry.AxisKind, entry.AxisID)
			continue
		}
		_ = r.store.Relay().MarkRebalanceDone(entry.ID)
	}
	return nil
}

func (r *Runner) compensateLogs(ctx context.Context) error {
	if r.client == nil {
		return nil
	}
	lastID, err := r.store.Relay().GetLastLogID()
	if err != nil {
		return err
	}
	logs, err := r.client.ListLogs(ctx, newapi.ListLogsParams{Page: 1, PageSize: 100, StartID: lastID})
	if err != nil {
		return err
	}
	for _, logEntry := range logs {
		payload := newapi.WebhookLogPayload{
			ID:        logEntry.ID,
			TokenID:   logEntry.TokenID,
			Quota:     logEntry.Quota,
			Model:     newapi.LogEntryModel(logEntry),
			CreatedAt: logEntry.CreatedAt,
		}
		if err := r.ingest.Ingest(ctx, payload); err != nil {
			r.logger.Warn("log compensation ingest failed", "log_id", logEntry.ID, "error", err)
		}
	}
	return nil
}

func backoff(attempts int) time.Duration {
	seconds := math.Min(300, math.Pow(2, float64(attempts)))
	return time.Duration(seconds) * time.Second
}
