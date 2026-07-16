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

func (r *logRepo) GetConsumeLogsByIDs(ctx context.Context, logIDs []int64) ([]store.RawConsumeLog, error) {
	if len(logIDs) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf(`
		SELECT id, token_id, quota, model_name, created_at, prompt_tokens, completion_tokens, use_time, content
		FROM %s
		WHERE id = ANY($1) AND type = $2 AND token_id > 0
	`, r.tables.logs)
	rows, err := r.db.Query(ctx, query, logIDs, store.NewAPILogTypeConsume)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]store.RawConsumeLog, 0, len(logIDs))
	for rows.Next() {
		var raw store.RawConsumeLog
		if err := rows.Scan(
			&raw.ID, &raw.TokenID, &raw.Quota, &raw.ModelName, &raw.CreatedAt,
			&raw.PromptTokens, &raw.CompletionTokens, &raw.UseTime, &raw.Content,
		); err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	return out, rows.Err()
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
