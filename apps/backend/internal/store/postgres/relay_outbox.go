package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *relayRepo) EnqueueRelayOutbox(ctx context.Context, entry store.RelayOutboxEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO relay_outbox (id, kind, payload, status, attempts, next_retry, last_error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`, entry.ID, entry.Kind, entry.Payload, entry.Status, entry.Attempts, entry.NextRetry, entry.LastError, entry.CreatedAt)
	return err
}

func (r *relayRepo) ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]store.RelayOutboxEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, kind, payload, status, attempts, next_retry, last_error, created_at
		FROM relay_outbox
		WHERE status = $1 AND next_retry <= NOW()
		ORDER BY created_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, store.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRelayOutbox(rows)
}

func scanRelayOutbox(rows pgx.Rows) ([]store.RelayOutboxEntry, error) {
	out := make([]store.RelayOutboxEntry, 0)
	for rows.Next() {
		var e store.RelayOutboxEntry
		if err := rows.Scan(&e.ID, &e.Kind, &e.Payload, &e.Status, &e.Attempts, &e.NextRetry, &e.LastError, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkRelayOutboxDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE relay_outbox SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}

func (r *relayRepo) MarkRelayOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE relay_outbox SET attempts = attempts + 1, next_retry = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, nextRetry, lastError)
	return err
}

func (r *relayRepo) HasIngestedLogID(ctx context.Context, logID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM ingested_log_ids WHERE log_id = $1)
	`, logID).Scan(&exists)
	return exists, err
}

func (r *relayRepo) InsertIngestedLogID(ctx context.Context, logID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ingested_log_ids (log_id) VALUES ($1) ON CONFLICT DO NOTHING
	`, logID)
	return err
}

func (r *relayRepo) GetLastLogID(ctx context.Context) (int64, error) {
	companyID := store.CompanyID(ctx)
	var id int64
	err := r.db.QueryRow(ctx, `SELECT last_log_id FROM relay_sync_cursors WHERE company_id = $1`, companyID).Scan(&id)
	return id, err
}

func (r *relayRepo) SetLastLogID(ctx context.Context, logID int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO relay_sync_cursors (company_id, last_log_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (company_id) DO UPDATE SET last_log_id = EXCLUDED.last_log_id, updated_at = NOW()
	`, companyID, logID)
	return err
}
