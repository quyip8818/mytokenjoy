package memory

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryRelayRepo) EnqueueWebhookOutbox(ctx context.Context, entry store.WebhookOutboxEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.webhookOutbox = append(r.data.webhookOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingWebhookOutbox(ctx context.Context, limit int) ([]store.WebhookOutboxEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]store.WebhookOutboxEntry, 0, limit)
	for i := range r.data.webhookOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.webhookOutbox[i]
		if e.Status == store.OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Attempts++
			r.data.webhookOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.webhookOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}
