package seed

import (
	"context"
	"encoding/json"

	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
)

func insertAudit(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO audit_settings (company_id, content_retention_enabled)
		VALUES ($1, $2) ON CONFLICT (company_id) DO NOTHING
	`, tid, snap.AuditSettings.ContentRetentionEnabled); err != nil {
		return err
	}
	for _, log := range snap.OperationLogs {
		createdAt, err := pkgtime.Parse(log.CreatedAt)
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
	for _, entry := range snap.UsageLedger {
		detailJSON, err := json.Marshal(entry.CallDetail)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO usage_ledger (
				id, company_id, event_type, idempotency_key, amount_cny,
				department_id, member_id, budget_group_id, platform_key_id,
				source, occurred_at, model, input_tokens, output_tokens,
				call_detail, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (company_id, id) DO NOTHING
		`, entry.ID, tid, entry.EventType, entry.IdempotencyKey, entry.AmountCNY,
			entry.DepartmentID, entry.MemberID, entry.BudgetGroupID, entry.PlatformKeyID,
			entry.Source, entry.OccurredAt.UTC(), entry.Model, entry.InputTokens, entry.OutputTokens,
			detailJSON, entry.CreatedAt.UTC()); err != nil {
			return err
		}
	}
	return nil
}
