package postgres

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) Approvals() []types.KeyApproval {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, type, applicant, applicant_id, department, reason, requested_quota,
			status, approver, reject_reason, created_at, resolved_at
		FROM key_approvals ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.KeyApproval, 0)
	for rows.Next() {
		var item types.KeyApproval
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Type, &item.Applicant, &item.ApplicantID, &item.Department,
			&item.Reason, &item.RequestedQuota, &item.Status, &item.Approver, &item.RejectReason,
			&createdAt, &resolvedAt,
		); err != nil {
			return nil
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		if resolvedAt != nil {
			s := formatSyncLogTime(*resolvedAt)
			item.ResolvedAt = &s
		}
		modelRows, err := r.db.Query(r.ctx, `
			SELECT model_name FROM key_approval_models WHERE approval_id = $1 ORDER BY model_name
		`, item.ID)
		if err == nil {
			for modelRows.Next() {
				var modelName string
				if err := modelRows.Scan(&modelName); err == nil {
					item.RequestedModels = append(item.RequestedModels, modelName)
				}
			}
			modelRows.Close()
		}
		items = append(items, item)
	}
	return store.CloneApprovals(items)
}

func (r *pgKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	cloned := store.CloneApprovals(approvals)
	ids := make([]string, len(cloned))
	for i, approval := range cloned {
		ids[i] = approval.ID
		createdAt, err := parseAPITime(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := parseAPITime(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO key_approvals (
				id, type, applicant, applicant_id, department, reason, requested_quota,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				type = EXCLUDED.type,
				applicant = EXCLUDED.applicant,
				applicant_id = EXCLUDED.applicant_id,
				department = EXCLUDED.department,
				reason = EXCLUDED.reason,
				requested_quota = EXCLUDED.requested_quota,
				status = EXCLUDED.status,
				approver = EXCLUDED.approver,
				reject_reason = EXCLUDED.reject_reason,
				created_at = EXCLUDED.created_at,
				resolved_at = EXCLUDED.resolved_at
		`, approval.ID, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedQuota, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return fmt.Errorf("upsert approval %s: %w", approval.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM key_approval_models WHERE approval_id = $1`, approval.ID); err != nil {
			return err
		}
		for _, modelName := range approval.RequestedModels {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO key_approval_models (approval_id, model_name) VALUES ($1, $2)
			`, approval.ID, modelName); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM key_approval_models`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM key_approvals`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "key_approval_models", "approval_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "key_approvals", ids)
}
