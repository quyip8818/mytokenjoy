//go:build testhook

package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func MainPool(st store.Store) *pgxpool.Pool {
	pg, ok := st.(*Store)
	if !ok || pg.pool == nil {
		panic("MainPool requires postgres store")
	}
	return pg.pool
}

func LogPool(st store.Store) *pgxpool.Pool {
	pg, ok := st.(*Store)
	if !ok || pg.logPool == nil {
		panic("LogPool requires postgres store with ingest enabled")
	}
	return pg.logPool
}

func InsertConsumeLog(ctx context.Context, pool *pgxpool.Pool, raw store.RawConsumeLog) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO newapi.logs (
			id, user_id, created_at, type, content, token_id, model_name, quota,
			prompt_tokens, completion_tokens, use_time
		) VALUES ($1, 0, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			token_id = EXCLUDED.token_id,
			quota = EXCLUDED.quota,
			model_name = EXCLUDED.model_name,
			created_at = EXCLUDED.created_at,
			prompt_tokens = EXCLUDED.prompt_tokens,
			completion_tokens = EXCLUDED.completion_tokens,
			use_time = EXCLUDED.use_time,
			content = EXCLUDED.content
	`, raw.ID, raw.CreatedAt, store.NewAPILogTypeConsume, raw.Content, raw.TokenID, raw.ModelName, raw.Quota,
		raw.PromptTokens, raw.CompletionTokens, raw.UseTime)
	return err
}

func TruncateLogTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE newapi.logs, backend.ingest_failures, backend.reconcile_cursors RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO backend.reconcile_cursors (stream, last_log_id)
		VALUES ($1, 0)
		ON CONFLICT (stream) DO UPDATE SET last_log_id = 0, updated_at = NOW()
	`, store.ReconcileStreamNewAPIConsume)
	return err
}

func GetIngestFailureByLogID(ctx context.Context, pool *pgxpool.Pool, logID int64) (store.IngestFailure, bool, error) {
	row := pool.QueryRow(ctx, `
		SELECT id, log_id, source, error, status, attempts, next_retry, created_at, updated_at
		FROM backend.ingest_failures
		WHERE log_id = $1
	`, logID)
	var f store.IngestFailure
	err := row.Scan(&f.ID, &f.LogID, &f.Source, &f.Error, &f.Status, &f.Attempts, &f.NextRetry, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IngestFailure{}, false, nil
	}
	if err != nil {
		return store.IngestFailure{}, false, err
	}
	return f, true, nil
}

func GetRelayOutboxByID(ctx context.Context, pool *pgxpool.Pool, id string) (store.RelayOutboxEntry, bool, error) {
	row := pool.QueryRow(ctx, `
		SELECT id, kind, payload, status, attempts, next_retry, last_error, created_at
		FROM outbox
		WHERE id = $1 AND channel = $2
	`, id, store.OutboxChannelRelay)
	var e store.RelayOutboxEntry
	err := row.Scan(&e.ID, &e.Kind, &e.Payload, &e.Status, &e.Attempts, &e.NextRetry, &e.LastError, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.RelayOutboxEntry{}, false, nil
	}
	if err != nil {
		return store.RelayOutboxEntry{}, false, err
	}
	return e, true, nil
}

func ListPendingRelayOutbox(ctx context.Context, pool *pgxpool.Pool, kind string, limit int) ([]store.RelayOutboxEntry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, kind, payload, status, attempts, next_retry, last_error, created_at
		FROM outbox
		WHERE channel = $1 AND status = $2 AND ($3 = '' OR kind = $3)
		ORDER BY created_at
		LIMIT $4
	`, store.OutboxChannelRelay, store.OutboxStatusPending, kind, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func ListNotificationLogs(ctx context.Context, pool *pgxpool.Pool, companyID int64) ([]types.NotificationLogEntry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, channel, event_type, recipient, payload, status, COALESCE(error, '')
		FROM notification_log
		WHERE company_id = $1
		ORDER BY created_at
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]types.NotificationLogEntry, 0)
	for rows.Next() {
		var e types.NotificationLogEntry
		if err := rows.Scan(&e.ID, &e.Channel, &e.EventType, &e.Recipient, &e.Payload, &e.Status, &e.Error); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
