package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

//go:embed logs_schema.sql
var logsSchemaSQL string

//go:embed logs_schema_isolated.sql
var logsSchemaIsolatedSQL string

type logRepo struct {
	db     *pgxpool.Pool
	tables logTables
}

func newLogRepo(db *pgxpool.Pool, tables logTables) *logRepo {
	return &logRepo{db: db, tables: tables}
}

func (r *logRepo) GetConsumeLogByID(ctx context.Context, logID int64) (*store.RawConsumeLog, error) {
	query := fmt.Sprintf(`
		SELECT id, token_id, quota, model_name, created_at, prompt_tokens, completion_tokens, use_time, content
		FROM %s
		WHERE id = $1 AND type = $2 AND token_id > 0
	`, r.tables.logs)
	row := r.db.QueryRow(ctx, query, logID, store.NewAPILogTypeConsume)

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
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE id > $1 AND type = $2 AND token_id > 0
		ORDER BY id
		LIMIT $3
	`, r.tables.logs)
	rows, err := r.db.Query(ctx, query, afterID, store.NewAPILogTypeConsume, limit)
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
	query := fmt.Sprintf(`
		SELECT last_log_id FROM %s WHERE stream = $1
	`, r.tables.reconcileCursors)
	var last int64
	err := r.db.QueryRow(ctx, query, stream).Scan(&last)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return last, err
}

func (r *logRepo) SetReconcileCursor(ctx context.Context, stream string, logID int64) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (stream, last_log_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (stream) DO UPDATE SET last_log_id = EXCLUDED.last_log_id, updated_at = NOW()
	`, r.tables.reconcileCursors)
	_, err := r.db.Exec(ctx, query, stream, logID)
	return err
}

func (r *logRepo) EnqueuePending(ctx context.Context, logID int64, source string) error {
	id := store.IngestJobID(logID)
	query := fmt.Sprintf(`
		INSERT INTO %s (id, log_id, source, error, status, attempts, next_retry, created_at, updated_at)
		VALUES ($1, $2, $3, '', $4, 0, NOW(), NOW(), NOW())
		ON CONFLICT (log_id) DO UPDATE SET
			source = EXCLUDED.source,
			status = $4,
			attempts = CASE WHEN %s.status = $5 THEN 0 ELSE %s.attempts END,
			error = CASE WHEN %s.status = $5 THEN '' ELSE %s.error END,
			next_retry = NOW(),
			updated_at = NOW()
	`, r.tables.ingestJobs, r.tables.ingestJobs, r.tables.ingestJobs, r.tables.ingestJobs, r.tables.ingestJobs)
	_, err := r.db.Exec(ctx, query, id, logID, source, store.IngestJobStatusPending, store.IngestJobStatusDead)
	return err
}

func (r *logRepo) UpsertJob(ctx context.Context, job store.IngestJob) error {
	status := job.Status
	if status == "" {
		status = store.IngestJobStatusPending
	}
	var nextRetry any
	if !job.NextRetry.IsZero() {
		nextRetry = job.NextRetry
	}
	var createdAt any
	if !job.CreatedAt.IsZero() {
		createdAt = job.CreatedAt
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (id, log_id, source, error, status, attempts, next_retry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7::timestamptz, NOW()), COALESCE($8::timestamptz, NOW()), NOW())
		ON CONFLICT (log_id) DO UPDATE SET
			source = EXCLUDED.source,
			error = EXCLUDED.error,
			updated_at = NOW()
	`, r.tables.ingestJobs)
	_, err := r.db.Exec(ctx, query, job.ID, job.LogID, job.Source, job.Error, status, job.Attempts, nextRetry, createdAt)
	return err
}

func (r *logRepo) ClaimPendingJobs(ctx context.Context, limit int) ([]store.IngestJob, error) {
	leaseSecs := store.IngestJobClaimLease().Seconds()
	query := fmt.Sprintf(`
		WITH claimed AS (
			SELECT id
			FROM %s
			WHERE status = $1 AND next_retry <= NOW() AND attempts < $2
			ORDER BY next_retry
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE %s AS j
		SET next_retry = NOW() + ($4::double precision * INTERVAL '1 second'), updated_at = NOW()
		FROM claimed
		WHERE j.id = claimed.id
		RETURNING j.id, j.log_id, j.source, j.error, j.status, j.attempts, j.next_retry, j.created_at, j.updated_at
	`, r.tables.ingestJobs, r.tables.ingestJobs)
	rows, err := r.db.Query(ctx, query, store.IngestJobStatusPending, store.IngestJobMaxAttempts, limit, leaseSecs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIngestJobs(rows)
}

func scanIngestJobs(rows pgx.Rows) ([]store.IngestJob, error) {
	out := make([]store.IngestJob, 0)
	for rows.Next() {
		var job store.IngestJob
		if err := rows.Scan(
			&job.ID, &job.LogID, &job.Source, &job.Error, &job.Status,
			&job.Attempts, &job.NextRetry, &job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, job)
	}
	return out, rows.Err()
}

func (r *logRepo) MarkJobDone(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.tables.ingestJobs)
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *logRepo) MarkJobRetry(ctx context.Context, id string, delay time.Duration, errMsg string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET attempts = attempts + 1,
		    next_retry = NOW() + ($2::double precision * INTERVAL '1 second'),
		    error = $3,
		    updated_at = NOW()
		WHERE id = $1
	`, r.tables.ingestJobs)
	_, err := r.db.Exec(ctx, query, id, delay.Seconds(), errMsg)
	return err
}

func (r *logRepo) MarkJobDead(ctx context.Context, id string, errMsg string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $2, error = $3, updated_at = NOW()
		WHERE id = $1
	`, r.tables.ingestJobs)
	_, err := r.db.Exec(ctx, query, id, store.IngestJobStatusDead, errMsg)
	return err
}

func (r *logRepo) CountConsumeLogsAfter(ctx context.Context, afterID int64) (int64, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE id > $1 AND type = $2 AND token_id > 0
	`, r.tables.logs)
	var count int64
	err := r.db.QueryRow(ctx, query, afterID, store.NewAPILogTypeConsume).Scan(&count)
	return count, err
}

func (r *logRepo) CountPendingIngestJobs(ctx context.Context) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE status = $1 AND attempts < $2
	`, r.tables.ingestJobs)
	var count int
	err := r.db.QueryRow(ctx, query, store.IngestJobStatusPending, store.IngestJobMaxAttempts).Scan(&count)
	return count, err
}

func (r *logRepo) IngestLagSeconds(ctx context.Context, afterID int64) (int64, error) {
	query := fmt.Sprintf(`
		SELECT MIN(created_at)
		FROM %s
		WHERE id > $1 AND type = $2 AND token_id > 0
	`, r.tables.logs)
	var oldestCreatedAt *int64
	err := r.db.QueryRow(ctx, query, afterID, store.NewAPILogTypeConsume).Scan(&oldestCreatedAt)
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

func applyLogsSchema(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	sql := logsSchemaSQL
	if cfg.LogSchemaIsolated {
		sql = logsSchemaIsolatedSQL
	}
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("apply logs schema: %w", err)
	}
	return nil
}
