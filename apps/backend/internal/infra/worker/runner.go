package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
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
	syncCursor     store.IngestDedupRepository
	lifecycle      relay.Lifecycle
	ingest         domainbudget.Ingestor
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
		syncCursor:     relayRepo,
		client:         client,
		lifecycle:      lifecycle,
		ingest:         ingest,
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
	return r.syncSvc.RunScheduledSync(company.WithDefaultCompany(ctx, r.cfg.DefaultCompanyID))
}

// RunOnce executes one worker cycle (outbox + rebalance + log compensation).
func (r *Runner) RunOnce(ctx context.Context) { r.tick(ctx) }

func (r *Runner) workerCtx(ctx context.Context, companyID int64) context.Context {
	return company.WithDefaultCompany(ctx, companyID)
}

func (r *Runner) processRelayOutbox(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.relayOutbox.ClaimPendingRelayOutbox(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		var processErr error
		switch entry.Kind {
		case store.OutboxKindCreateToken:
			var payload relay.CreateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal create token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			_, processErr = r.lifecycle.TrySyncCreate(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindUpdateToken:
			var payload relay.UpdateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.lifecycle.SyncUpdatePlatformKey(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindRevokeToken:
			var payload relay.UpdateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal revoke token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.lifecycle.SyncRevokePlatformKey(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindUpsertChannel:
			var payload relay.UpsertChannelOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal upsert channel payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.lifecycle.SyncUpsertProviderKey(entryCtx, payload.ProviderKeyID)
		case store.OutboxKindUpdateModelLimits:
			var payload relay.UpdateModelLimitsOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update model limits payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.lifecycle.SyncModelLimitsForDepartment(entryCtx, payload.DepartmentID)
		case store.OutboxKindRebalanceToken:
			var payload relay.RebalanceAxisOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal rebalance payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.rebalance.ProcessAxis(entryCtx, payload.AxisKind, payload.AxisID)
		default:
			processErr = fmt.Errorf("unknown relay outbox kind: %s", entry.Kind)
		}
		if processErr != nil {
			next := time.Now().Add(backoff(entry.Attempts))
			r.markRelayOutboxRetry(workerCtx, entry.ID, next, processErr.Error())
			continue
		}
		r.markRelayOutboxDone(workerCtx, entry.ID)
	}
	return nil
}

func (r *Runner) processWebhookOutbox(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.webhookOutbox.ClaimPendingWebhookOutbox(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := r.ingest.IngestFromOutbox(workerCtx, entry.Payload); err != nil {
			next := time.Now().Add(backoff(entry.Attempts))
			r.markWebhookOutboxRetry(workerCtx, entry.ID, next, err.Error())
			continue
		}
		r.markWebhookOutboxDone(workerCtx, entry.ID)
	}
	return nil
}

func (r *Runner) processRebalance(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.rebalanceQueue.ClaimPendingRebalance(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryCtx := r.workerCtx(ctx, entry.CompanyID)
		if err := r.rebalance.ProcessAxis(entryCtx, entry.AxisKind, entry.AxisID); err != nil {
			r.logger.Warn("rebalance failed", "axis", entry.AxisKind, "id", entry.AxisID, "company_id", entry.CompanyID, "error", err)
			if enqueueErr := r.lifecycle.EnqueueRebalanceAxis(entryCtx, entry.AxisKind, entry.AxisID); enqueueErr != nil {
				r.logger.Warn("re-enqueue rebalance failed", "axis", entry.AxisKind, "id", entry.AxisID, "error", enqueueErr)
			}
			continue
		}
		r.markRebalanceDone(workerCtx, entry.ID)
	}
	return nil
}

func (r *Runner) compensateLogs(ctx context.Context) error {
	if r.client == nil {
		return nil
	}
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	lastID, err := r.syncCursor.GetLastLogID(workerCtx)
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
