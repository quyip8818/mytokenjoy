package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgAuditRepo struct {
	db dbQuerier
}

func (r *pgAuditRepo) Settings(ctx context.Context) (types.AuditSettings, error) {
	companyID := store.CompanyID(ctx)
	var enabled bool
	err := r.db.QueryRow(ctx, `
		SELECT content_retention_enabled FROM audit_settings WHERE company_id = $1
	`, companyID).Scan(&enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.AuditSettings{}, nil
		}
		return types.AuditSettings{}, err
	}
	return types.AuditSettings{ContentRetentionEnabled: enabled}, nil
}

func (r *pgAuditRepo) SetSettings(ctx context.Context, settings types.AuditSettings) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_settings (company_id, content_retention_enabled, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			content_retention_enabled = EXCLUDED.content_retention_enabled,
			updated_at = NOW()
	`, companyID, settings.ContentRetentionEnabled)
	if err != nil {
		return fmt.Errorf("upsert audit settings: %w", err)
	}
	return nil
}

func (r *pgAuditRepo) OperationLogs(ctx context.Context) ([]types.OperationLog, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, action, operator, operator_id, actor_type, target, detail, ip, created_at
		FROM operation_logs
		WHERE company_id = $1
		ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.OperationLog, 0)
	for rows.Next() {
		var item types.OperationLog
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Action, &item.Operator, &item.OperatorID, &item.ActorType,
			&item.Target, &item.Detail, &item.IP, &createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgAuditRepo) ListOperationsPage(ctx context.Context, filter store.AuditOperationFilter) ([]types.OperationLog, int, error) {
	companyID := store.CompanyID(ctx)
	where, args := buildAuditOperationWhere(companyID, filter)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM operation_logs WHERE %s`, where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizeAuditPage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	listArgs := append(append([]any{}, args...), pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, action, operator, operator_id, actor_type, target, detail, ip, created_at
		FROM operation_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]types.OperationLog, 0)
	for rows.Next() {
		var item types.OperationLog
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Action, &item.Operator, &item.OperatorID, &item.ActorType,
			&item.Target, &item.Detail, &item.IP, &createdAt,
		); err != nil {
			return nil, 0, err
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func buildAuditOperationWhere(companyID int64, filter store.AuditOperationFilter) (string, []any) {
	clauses := []string{"company_id = $1"}
	args := []any{companyID}
	idx := 2

	if filter.Action != "" {
		clauses = append(clauses, fmt.Sprintf("action = $%d", idx))
		args = append(args, filter.Action)
		idx++
	}
	if filter.OperatorID != "" {
		clauses = append(clauses, fmt.Sprintf("operator_id = $%d", idx))
		args = append(args, filter.OperatorID)
		idx++
	}
	if filter.From != "" {
		clauses = append(clauses, fmt.Sprintf("created_at::date >= $%d::date", idx))
		args = append(args, filter.From)
		idx++
	}
	if filter.To != "" {
		clauses = append(clauses, fmt.Sprintf("created_at::date <= $%d::date", idx))
		args = append(args, filter.To)
		idx++
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		clauses = append(clauses, fmt.Sprintf(`(
			detail ILIKE $%d OR target ILIKE $%d OR operator ILIKE $%d
		)`, idx, idx, idx))
		args = append(args, "%"+escapeLikePattern(kw)+"%")
		idx++
	}
	return strings.Join(clauses, " AND "), args
}

func normalizeAuditPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func (r *pgAuditRepo) AppendOperationLog(ctx context.Context, log types.OperationLog) error {
	companyID := store.CompanyID(ctx)
	createdAt, err := parseTimeOrNow(log.CreatedAt)
	if err != nil {
		return err
	}
	actorType := log.ActorType
	if actorType == "" {
		actorType = store.ActorTypeMember
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO operation_logs (id, company_id, action, operator, operator_id, actor_type, target, detail, ip, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (company_id, id, created_at) DO NOTHING
	`, log.ID, companyID, log.Action, log.Operator, log.OperatorID, actorType, log.Target, log.Detail, log.IP, createdAt)
	return err
}

var _ store.AuditRepository = (*pgAuditRepo)(nil)
