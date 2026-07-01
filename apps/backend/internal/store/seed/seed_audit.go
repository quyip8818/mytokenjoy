package seed

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/timeparse"
)

func insertAudit(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO audit_settings (company_id, content_retention_enabled)
		VALUES ($1, $2) ON CONFLICT (company_id) DO NOTHING
	`, tid, snap.AuditSettings.ContentRetentionEnabled); err != nil {
		return err
	}
	for _, log := range snap.OperationLogs {
		createdAt, err := timeparse.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		actorType := log.ActorType
		if actorType == "" {
			actorType = store.ActorTypeMember
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO operation_logs (id, company_id, action, operator, operator_id, actor_type, target, detail, ip, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (company_id, id) DO NOTHING
		`, log.ID, tid, log.Action, log.Operator, log.OperatorID, actorType, log.Target, log.Detail, log.IP, createdAt); err != nil {
			return err
		}
	}
	for _, log := range snap.CallLogs {
		createdAt, err := timeparse.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO call_logs (
				id, company_id, caller, caller_id, caller_type, model, provider,
				input_tokens, output_tokens, latency_ms, status, cost,
				input_preview, output_preview, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (company_id, id) DO NOTHING
		`, log.ID, tid, log.Caller, log.CallerID, log.CallerType, log.Model, log.Provider,
			log.InputTokens, log.OutputTokens, log.LatencyMs, log.Status, log.Cost,
			log.InputPreview, log.OutputPreview, createdAt); err != nil {
			return err
		}
	}
	return nil
}
