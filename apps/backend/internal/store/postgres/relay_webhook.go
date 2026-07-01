package postgres

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

func (r *relayRepo) EnqueueWebhookOutbox(ctx context.Context, entry store.WebhookOutboxEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO webhook_outbox (id, payload, status, attempts, next_retry, last_error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
	`, entry.ID, entry.Payload, entry.Status, entry.Attempts, entry.NextRetry, entry.LastError, entry.CreatedAt)
	return err
}

func (r *relayRepo) ClaimPendingWebhookOutbox(ctx context.Context, limit int) ([]store.WebhookOutboxEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, payload, status, attempts, next_retry, last_error, created_at
		FROM webhook_outbox
		WHERE status = $1 AND next_retry <= NOW()
		ORDER BY created_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, store.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.WebhookOutboxEntry, 0)
	for rows.Next() {
		var e store.WebhookOutboxEntry
		if err := rows.Scan(&e.ID, &e.Payload, &e.Status, &e.Attempts, &e.NextRetry, &e.LastError, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkWebhookOutboxDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_outbox SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}

func (r *relayRepo) MarkWebhookOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_outbox SET attempts = attempts + 1, next_retry = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, nextRetry, lastError)
	return err
}
