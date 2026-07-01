package worker

import (
	"context"
	"time"
)

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
