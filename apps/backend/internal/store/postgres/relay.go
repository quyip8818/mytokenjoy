package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/store"
)

type dbQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type relayRepo struct {
	db  dbQuerier
	ctx context.Context
}

func newRelayRepo(ctx context.Context, db dbQuerier) *relayRepo {
	if ctx == nil {
		ctx = context.Background()
	}
	return &relayRepo{db: db, ctx: ctx}
}

func scanMapping(row pgx.Row) (store.RelayMapping, error) {
	var m store.RelayMapping
	var memberID, budgetGroupID *string
	var tokenID, remainQuota *int64
	var syncedAt *time.Time
	err := row.Scan(
		&m.PlatformKeyID, &tokenID, &memberID, &m.DepartmentID, &budgetGroupID,
		&m.RelayGroup, &m.SyncStatus, &syncedAt, &remainQuota,
	)
	if err != nil {
		return store.RelayMapping{}, err
	}
	m.NewAPITokenID = tokenID
	m.MemberID = memberID
	m.BudgetGroupID = budgetGroupID
	m.SyncedAt = syncedAt
	m.RelayRemainQuota = remainQuota
	return m, nil
}

const mappingSelect = `
	SELECT platform_key_id, newapi_token_id, member_id, department_id, budget_group_id,
	       relay_group, sync_status, synced_at, relay_remain_quota
	FROM relay_mappings
`

func (r *relayRepo) GetMappingByPlatformKeyID(platformKeyID string) (*store.RelayMapping, error) {
	row := r.db.QueryRow(r.ctx, mappingSelect+` WHERE platform_key_id = $1`, platformKeyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByNewAPITokenID(tokenID int64) (*store.RelayMapping, error) {
	row := r.db.QueryRow(r.ctx, mappingSelect+` WHERE newapi_token_id = $1`, tokenID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) listMappings(where string, arg any) ([]store.RelayMapping, error) {
	query := mappingSelect
	if where != "" {
		query += " WHERE " + where
	}
	rows, err := r.db.Query(r.ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.RelayMapping, 0)
	for rows.Next() {
		m, err := scanMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *relayRepo) ListMappingsByMemberID(memberID string) ([]store.RelayMapping, error) {
	return r.listMappings("member_id = $1", memberID)
}

func (r *relayRepo) ListMappingsByDepartmentID(departmentID string) ([]store.RelayMapping, error) {
	return r.listMappings("department_id = $1", departmentID)
}

func (r *relayRepo) ListMappingsByBudgetGroupID(budgetGroupID string) ([]store.RelayMapping, error) {
	return r.listMappings("budget_group_id = $1", budgetGroupID)
}

func (r *relayRepo) ListActiveMappings() ([]store.RelayMapping, error) {
	return r.listMappings("sync_status = $1", store.RelaySyncStatusSynced)
}

func (r *relayRepo) UpsertMapping(mapping store.RelayMapping) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO relay_mappings (
			platform_key_id, newapi_token_id, member_id, department_id, budget_group_id,
			relay_group, sync_status, synced_at, relay_remain_quota, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (platform_key_id) DO UPDATE SET
			newapi_token_id = EXCLUDED.newapi_token_id,
			member_id = EXCLUDED.member_id,
			department_id = EXCLUDED.department_id,
			budget_group_id = EXCLUDED.budget_group_id,
			relay_group = EXCLUDED.relay_group,
			sync_status = EXCLUDED.sync_status,
			synced_at = EXCLUDED.synced_at,
			relay_remain_quota = EXCLUDED.relay_remain_quota,
			updated_at = NOW()
	`, mapping.PlatformKeyID, mapping.NewAPITokenID, mapping.MemberID, mapping.DepartmentID,
		mapping.BudgetGroupID, mapping.RelayGroup, mapping.SyncStatus, mapping.SyncedAt, mapping.RelayRemainQuota)
	return err
}

func (r *relayRepo) UpdateMappingSync(
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE relay_mappings
		SET newapi_token_id = $2, sync_status = $3, relay_remain_quota = $4,
		    synced_at = $5, updated_at = NOW()
		WHERE platform_key_id = $1
	`, platformKeyID, tokenID, status, remainQuota, syncedAt)
	return err
}

func (r *relayRepo) UpdateMappingRemainQuota(platformKeyID string, remainQuota int64) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE relay_mappings SET relay_remain_quota = $2, updated_at = NOW()
		WHERE platform_key_id = $1
	`, platformKeyID, remainQuota)
	return err
}

func (r *relayRepo) EnqueueRelayOutbox(entry store.RelayOutboxEntry) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO relay_outbox (id, kind, payload, status, attempts, next_retry, last_error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`, entry.ID, entry.Kind, entry.Payload, entry.Status, entry.Attempts, entry.NextRetry, entry.LastError, entry.CreatedAt)
	return err
}

func (r *relayRepo) ClaimPendingRelayOutbox(limit int) ([]store.RelayOutboxEntry, error) {
	rows, err := r.db.Query(r.ctx, `
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

func (r *relayRepo) MarkRelayOutboxDone(id string) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE relay_outbox SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}

func (r *relayRepo) MarkRelayOutboxRetry(id string, nextRetry time.Time, lastError string) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE relay_outbox SET attempts = attempts + 1, next_retry = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, nextRetry, lastError)
	return err
}

func (r *relayRepo) EnqueueWebhookOutbox(entry store.WebhookOutboxEntry) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO webhook_outbox (id, payload, status, attempts, next_retry, last_error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
	`, entry.ID, entry.Payload, entry.Status, entry.Attempts, entry.NextRetry, entry.LastError, entry.CreatedAt)
	return err
}

func (r *relayRepo) ClaimPendingWebhookOutbox(limit int) ([]store.WebhookOutboxEntry, error) {
	rows, err := r.db.Query(r.ctx, `
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

func (r *relayRepo) MarkWebhookOutboxDone(id string) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE webhook_outbox SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}

func (r *relayRepo) MarkWebhookOutboxRetry(id string, nextRetry time.Time, lastError string) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE webhook_outbox SET attempts = attempts + 1, next_retry = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, nextRetry, lastError)
	return err
}

func (r *relayRepo) HasIngestedLogID(logID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(r.ctx, `
		SELECT EXISTS (SELECT 1 FROM ingested_log_ids WHERE log_id = $1)
	`, logID).Scan(&exists)
	return exists, err
}

func (r *relayRepo) InsertIngestedLogID(logID int64) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO ingested_log_ids (log_id) VALUES ($1) ON CONFLICT DO NOTHING
	`, logID)
	return err
}

func (r *relayRepo) GetLastLogID() (int64, error) {
	var id int64
	err := r.db.QueryRow(r.ctx, `SELECT last_log_id FROM relay_sync_cursors WHERE id = 1`).Scan(&id)
	return id, err
}

func (r *relayRepo) SetLastLogID(logID int64) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE relay_sync_cursors SET last_log_id = $1, updated_at = NOW() WHERE id = 1
	`, logID)
	return err
}

func (r *relayRepo) EnqueueRebalance(axisKind, axisID string) error {
	id := fmt.Sprintf("rb-%s-%s-%d", axisKind, axisID, time.Now().UnixNano())
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO rebalance_queue (id, axis_kind, axis_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (axis_kind, axis_id, status) DO NOTHING
	`, id, axisKind, axisID, store.OutboxStatusPending)
	return err
}

func (r *relayRepo) ClaimPendingRebalance(limit int) ([]store.RebalanceQueueEntry, error) {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, axis_kind, axis_id, status
		FROM rebalance_queue
		WHERE status = $1
		ORDER BY created_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, store.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.RebalanceQueueEntry, 0)
	for rows.Next() {
		var e store.RebalanceQueueEntry
		if err := rows.Scan(&e.ID, &e.AxisKind, &e.AxisID, &e.Status); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkRebalanceDone(id string) error {
	_, err := r.db.Exec(r.ctx, `
		UPDATE rebalance_queue SET status = $2 WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}
