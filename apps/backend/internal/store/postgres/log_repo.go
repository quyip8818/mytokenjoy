package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
)

//go:embed logs_schema.sql
var logsSchemaSQL string

type logRepo struct {
	db *pgxpool.Pool
}

func newLogRepo(db *pgxpool.Pool) *logRepo {
	return &logRepo{db: db}
}

func (r *logRepo) GetConsumeLogByID(ctx context.Context, logID int64) (*store.RawConsumeLog, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, token_id, quota, model_name, created_at, prompt_tokens, completion_tokens, use_time, content
		FROM newapi.logs
		WHERE id = $1 AND type = $2 AND token_id > 0
	`, logID, store.NewAPILogTypeConsume)

	var raw store.RawConsumeLog
	err := row.Scan(
		&raw.ID, &raw.TokenID, &raw.Quota, &raw.ModelName, &raw.CreatedAt,
		&raw.PromptTokens, &raw.CompletionTokens, &raw.UseTime, &raw.Content,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrConsumeLogNotFound
	}
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func (r *logRepo) ListConsumeLogIDsAfter(ctx context.Context, afterID int64, limit int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id
		FROM newapi.logs
		WHERE id > $1 AND type = $2 AND token_id > 0
		ORDER BY id
		LIMIT $3
	`, afterID, store.NewAPILogTypeConsume, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *logRepo) GetReconcileCursor(ctx context.Context, stream string) (int64, error) {
	var last int64
	err := r.db.QueryRow(ctx, `
		SELECT last_log_id FROM backend.reconcile_cursors WHERE stream = $1
	`, stream).Scan(&last)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return last, err
}

func (r *logRepo) SetReconcileCursor(ctx context.Context, stream string, logID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO backend.reconcile_cursors (stream, last_log_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (stream) DO UPDATE SET last_log_id = EXCLUDED.last_log_id, updated_at = NOW()
	`, stream, logID)
	return err
}

func (r *logRepo) UpsertFailure(ctx context.Context, f store.IngestFailure) error {
	status := f.Status
	if status == "" {
		status = store.IngestFailureStatusPending
	}
	nextRetry := f.NextRetry
	if nextRetry.IsZero() {
		nextRetry = time.Now()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO backend.ingest_failures (id, log_id, source, error, status, attempts, next_retry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW()), NOW())
		ON CONFLICT (log_id) DO UPDATE SET
			source = EXCLUDED.source,
			error = EXCLUDED.error,
			updated_at = NOW()
	`, f.ID, f.LogID, f.Source, f.Error, status, f.Attempts, nextRetry, f.CreatedAt)
	return err
}

func (r *logRepo) ClaimPendingFailures(ctx context.Context, limit int) ([]store.IngestFailure, error) {
	leaseUntil := time.Now().Add(store.FailureClaimLease())
	rows, err := r.db.Query(ctx, `
		WITH claimed AS (
			SELECT id
			FROM backend.ingest_failures
			WHERE status = $1 AND next_retry <= NOW() AND attempts < $2
			ORDER BY next_retry
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE backend.ingest_failures AS f
		SET next_retry = $4, updated_at = NOW()
		FROM claimed
		WHERE f.id = claimed.id
		RETURNING f.id, f.log_id, f.source, f.error, f.status, f.attempts, f.next_retry, f.created_at, f.updated_at
	`, store.IngestFailureStatusPending, store.IngestFailureMaxAttempts, limit, leaseUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIngestFailures(rows)
}

func scanIngestFailures(rows pgx.Rows) ([]store.IngestFailure, error) {
	out := make([]store.IngestFailure, 0)
	for rows.Next() {
		var f store.IngestFailure
		if err := rows.Scan(
			&f.ID, &f.LogID, &f.Source, &f.Error, &f.Status,
			&f.Attempts, &f.NextRetry, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *logRepo) MarkFailureDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM backend.ingest_failures WHERE id = $1`, id)
	return err
}

func (r *logRepo) MarkFailureRetry(ctx context.Context, id string, next time.Time, errMsg string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE backend.ingest_failures
		SET attempts = attempts + 1, next_retry = $2, error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, next, errMsg)
	return err
}

func (r *logRepo) MarkFailureDead(ctx context.Context, id string, errMsg string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE backend.ingest_failures
		SET status = $2, error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, store.IngestFailureStatusDead, errMsg)
	return err
}

func (r *logRepo) CountConsumeLogsAfter(ctx context.Context, afterID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM newapi.logs
		WHERE id > $1 AND type = $2 AND token_id > 0
	`, afterID, store.NewAPILogTypeConsume).Scan(&count)
	return count, err
}

func (r *logRepo) CountPendingIngestFailures(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM backend.ingest_failures
		WHERE status = $1 AND attempts < $2
	`, store.IngestFailureStatusPending, store.IngestFailureMaxAttempts).Scan(&count)
	return count, err
}

func (r *logRepo) IngestLagSeconds(ctx context.Context, afterID int64) (int64, error) {
	var oldestCreatedAt *int64
	err := r.db.QueryRow(ctx, `
		SELECT MIN(created_at)
		FROM newapi.logs
		WHERE id > $1 AND type = $2 AND token_id > 0
	`, afterID, store.NewAPILogTypeConsume).Scan(&oldestCreatedAt)
	if err != nil || oldestCreatedAt == nil {
		return 0, err
	}
	now := time.Now().Unix()
	lag := now - *oldestCreatedAt
	if lag < 0 {
		return 0, nil
	}
	return lag, nil
}

func applyLogsSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, logsSchemaSQL); err != nil {
		return fmt.Errorf("apply logs schema: %w", err)
	}
	return nil
}
