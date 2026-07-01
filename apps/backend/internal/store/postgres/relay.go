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
	db dbQuerier
}

func newRelayRepo(db dbQuerier) *relayRepo {
	return &relayRepo{db: db}
}

func scanMapping(row pgx.Row) (store.RelayMapping, error) {
	var m store.RelayMapping
	var memberID, budgetGroupID *string
	var tokenID, remainQuota *int64
	var syncedAt *time.Time
	err := row.Scan(
		&m.CompanyID, &m.PlatformKeyID, &tokenID, &memberID, &m.DepartmentID, &budgetGroupID,
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
	SELECT company_id, platform_key_id, newapi_token_id, member_id, department_id, budget_group_id,
	       relay_group, sync_status, synced_at, relay_remain_quota
	FROM relay_mappings
`

func (r *relayRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE company_id = $1 AND platform_key_id = $2`, companyID, platformKeyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByFullKey(ctx context.Context, fullKey string) (*store.RelayMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+`
		WHERE (company_id, platform_key_id) IN (
			SELECT company_id, id FROM platform_keys WHERE full_key = $1
		)
	`, fullKey)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE company_id = $1 AND newapi_token_id = $2`, companyID, tokenID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE newapi_token_id = $1`, tokenID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) listMappings(ctx context.Context, where string, args ...any) ([]store.RelayMapping, error) {
	query := mappingSelect
	if where != "" {
		query += " WHERE " + where
	}
	rows, err := r.db.Query(ctx, query, args...)
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

func (r *relayRepo) ListMappingsByMemberID(ctx context.Context, memberID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND member_id = $2", companyID, memberID)
}

func (r *relayRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND department_id = $2", companyID, departmentID)
}

func (r *relayRepo) ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND budget_group_id = $2", companyID, budgetGroupID)
}

func (r *relayRepo) ListActiveMappings(ctx context.Context) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.RelayMapping, error) {
	return r.listMappings(ctx, "company_id = $1 AND sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) UpsertMapping(ctx context.Context, mapping store.RelayMapping) error {
	companyID := store.CompanyID(ctx)
	if mapping.CompanyID > 0 {
		companyID = mapping.CompanyID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO relay_mappings (
			platform_key_id, company_id, newapi_token_id, member_id, department_id, budget_group_id,
			relay_group, sync_status, synced_at, relay_remain_quota, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
		ON CONFLICT (company_id, platform_key_id) DO UPDATE SET
			newapi_token_id = EXCLUDED.newapi_token_id,
			member_id = EXCLUDED.member_id,
			department_id = EXCLUDED.department_id,
			budget_group_id = EXCLUDED.budget_group_id,
			relay_group = EXCLUDED.relay_group,
			sync_status = EXCLUDED.sync_status,
			synced_at = EXCLUDED.synced_at,
			relay_remain_quota = EXCLUDED.relay_remain_quota,
			updated_at = NOW()
	`, mapping.PlatformKeyID, companyID, mapping.NewAPITokenID, mapping.MemberID, mapping.DepartmentID,
		mapping.BudgetGroupID, mapping.RelayGroup, mapping.SyncStatus, mapping.SyncedAt, mapping.RelayRemainQuota)
	return err
}

func (r *relayRepo) UpdateMappingSync(
	ctx context.Context,
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE relay_mappings
		SET newapi_token_id = $3, sync_status = $4, relay_remain_quota = $5,
		    synced_at = $6, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, tokenID, status, remainQuota, syncedAt)
	return err
}

func (r *relayRepo) UpdateMappingRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE relay_mappings SET relay_remain_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, remainQuota)
	return err
}

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

func (r *relayRepo) EnqueueRebalance(ctx context.Context, axisKind, axisID string) error {
	companyID := store.CompanyID(ctx)
	id := fmt.Sprintf("rb-%d-%s-%s-%d", companyID, axisKind, axisID, time.Now().UnixNano())
	_, err := r.db.Exec(ctx, `
		INSERT INTO rebalance_queue (id, company_id, axis_kind, axis_id, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (company_id, axis_kind, axis_id, status) DO NOTHING
	`, id, companyID, axisKind, axisID, store.OutboxStatusPending)
	return err
}

func (r *relayRepo) ClaimPendingRebalance(ctx context.Context, limit int) ([]store.RebalanceQueueEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, axis_kind, axis_id, status
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
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.AxisKind, &e.AxisID, &e.Status); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkRebalanceDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE rebalance_queue SET status = $2 WHERE id = $1
	`, id, store.OutboxStatusDone)
	return err
}

var _ store.RelayRepository = (*relayRepo)(nil)
