//go:build testhook

package postgres

import (
	"context"
	"errors"
	"fmt"

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

func logTablesFromStore(st store.Store) logTables {
	pg, ok := st.(*Store)
	if !ok || pg.logPool == nil {
		panic("log tables require postgres store with ingest enabled")
	}
	return pg.logTables
}

func InsertConsumeLog(ctx context.Context, st store.Store, raw store.RawConsumeLog) error {
	tables := logTablesFromStore(st)
	pool := LogPool(st)
	query := fmt.Sprintf(`
		INSERT INTO %s (
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
	`, tables.logs)
	_, err := pool.Exec(ctx, query, raw.ID, raw.CreatedAt, store.NewAPILogTypeConsume, raw.Content, raw.TokenID, raw.ModelName, raw.Quota,
		raw.PromptTokens, raw.CompletionTokens, raw.UseTime)
	return err
}

func GetRiverJobByID(ctx context.Context, pool *pgxpool.Pool, id int64) (store.RiverJobView, bool, error) {
	row := pool.QueryRow(ctx, `
		SELECT id::text, kind, args, state::text,
			CASE WHEN cardinality(errors) > 0 THEN (errors[cardinality(errors)]->>'error') ELSE NULL END
		FROM river_job
		WHERE id = $1
	`, id)
	var e store.RiverJobView
	var jobID string
	err := row.Scan(&jobID, &e.Kind, &e.Payload, &e.Status, &e.LastError)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.RiverJobView{}, false, nil
	}
	if err != nil {
		return store.RiverJobView{}, false, err
	}
	e.ID = jobID
	return e, true, nil
}

func ListPendingRiverJobs(ctx context.Context, pool *pgxpool.Pool, kind, subKind string, limit int) (int, error) {
	rows, err := pool.Query(ctx, `
		SELECT 1 FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND ($2 = '' OR args->>'sub_kind' = $2)
		LIMIT $3
	`, kind, subKind, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n, rows.Err()
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
