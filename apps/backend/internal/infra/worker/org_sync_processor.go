package worker

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func (r *Runner) processOrgSync(ctx context.Context) error {
	if r.syncSvc == nil {
		return nil
	}
	return r.syncSvc.RunScheduledSync(company.WithDefaultCompany(ctx, r.cfg.DefaultCompanyID))
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
