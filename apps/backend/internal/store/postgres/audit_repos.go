package postgres

import (
	"context"
	"fmt"
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
		SELECT id, action, operator, operator_id, COALESCE(actor_type, 'member'), target, detail, ip, created_at
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
	return store.CloneOperationLogs(items), nil
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
		ON CONFLICT (company_id, id) DO NOTHING
	`, log.ID, companyID, log.Action, log.Operator, log.OperatorID, actorType, log.Target, log.Detail, log.IP, createdAt)
	return err
}

func (r *pgAuditRepo) CallLogs(ctx context.Context) ([]types.CallLog, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, caller, caller_id, caller_type, model, provider,
			input_tokens, output_tokens, latency_ms, status, cost,
			input_preview, output_preview, created_at
		FROM call_logs
		WHERE company_id = $1
		ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.CallLog, 0)
	for rows.Next() {
		var item types.CallLog
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Caller, &item.CallerID, &item.CallerType,
			&item.Model, &item.Provider,
			&item.InputTokens, &item.OutputTokens, &item.LatencyMs,
			&item.Status, &item.Cost,
			&item.InputPreview, &item.OutputPreview, &createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneCallLogs(items), nil
}
