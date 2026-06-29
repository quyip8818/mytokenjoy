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
	ctx context.Context
	db  dbQuerier
}

func (r *pgAuditRepo) Settings() types.AuditSettings {
	var enabled bool
	err := r.db.QueryRow(r.ctx, `
		SELECT content_retention_enabled FROM audit_settings WHERE id = 1
	`).Scan(&enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.AuditSettings{}
		}
		return types.AuditSettings{}
	}
	return types.AuditSettings{ContentRetentionEnabled: enabled}
}

func (r *pgAuditRepo) SetSettings(settings types.AuditSettings) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO audit_settings (id, content_retention_enabled, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE SET
			content_retention_enabled = EXCLUDED.content_retention_enabled,
			updated_at = NOW()
	`, settings.ContentRetentionEnabled)
	if err != nil {
		return fmt.Errorf("upsert audit settings: %w", err)
	}
	return nil
}

func (r *pgAuditRepo) OperationLogs() []types.OperationLog {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, action, operator, operator_id, target, detail, ip, created_at
		FROM operation_logs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	items := make([]types.OperationLog, 0)
	for rows.Next() {
		var item types.OperationLog
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Action, &item.Operator, &item.OperatorID,
			&item.Target, &item.Detail, &item.IP, &createdAt,
		); err != nil {
			return nil
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		items = append(items, item)
	}
	return store.CloneOperationLogs(items)
}

func (r *pgAuditRepo) CallLogs() []types.CallLog {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, caller, caller_id, caller_type, model, provider,
			input_tokens, output_tokens, latency_ms, status, cost,
			input_preview, output_preview, created_at
		FROM call_logs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
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
			return nil
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		items = append(items, item)
	}
	return store.CloneCallLogs(items)
}
