package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) BudgetApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT ba.id, ba.applicant_id, ba.applicant_name, ba.department_name, ba.amount, ba.reason,
			ba.status, ba.reject_reason, ba.created_at, ba.resolved_at,
			m.department_id
		FROM budget_approvals ba
		LEFT JOIN members m ON m.company_id = ba.company_id AND m.id = ba.applicant_id
		WHERE ba.company_id = $1 ORDER BY ba.created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.BudgetApproval, 0)
	for rows.Next() {
		var item types.BudgetApproval
		var applicantID *uuid.UUID
		var departmentID *uuid.UUID
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(
			&item.ID, &applicantID, &item.ApplicantName, &item.DepartmentName,
			&item.Amount, &item.Reason, &item.Status, &item.RejectReason,
			&createdAt, &resolvedAt, &departmentID,
		); err != nil {
			return nil, err
		}
		if applicantID != nil {
			item.ApplicantID = *applicantID
		}
		if departmentID != nil {
			item.DepartmentID = *departmentID
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		if resolvedAt != nil {
			s := formatSyncLogTime(*resolvedAt)
			item.ResolvedAt = &s
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgBudgetRepo) SetBudgetApprovals(ctx context.Context, items []types.BudgetApproval) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneBudgetApprovals(items)
	ids := make([]uuid.UUID, len(cloned))
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
		var applicantID *uuid.UUID
		if approval.ApplicantID != uuid.Nil {
			applicantID = &approval.ApplicantID
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_approvals (
				id, company_id, applicant_id, applicant_name, department_name,
				amount, reason, status, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (company_id, id) DO UPDATE SET
				applicant_id = EXCLUDED.applicant_id,
				applicant_name = EXCLUDED.applicant_name,
				department_name = EXCLUDED.department_name,
				amount = EXCLUDED.amount,
				reason = EXCLUDED.reason,
				status = EXCLUDED.status,
				reject_reason = EXCLUDED.reject_reason,
				created_at = EXCLUDED.created_at,
				resolved_at = EXCLUDED.resolved_at
		`, approval.ID, companyID, applicantID, approval.ApplicantName, approval.DepartmentName,
			approval.Amount, approval.Reason, approval.Status, approval.RejectReason,
			createdAt, resolvedAt); err != nil {
			return fmt.Errorf("upsert budget approval %s: %w", approval.ID, err)
		}
	}
	if len(ids) == 0 {
		_, err := r.db.Exec(ctx, `DELETE FROM budget_approvals WHERE company_id = $1`, companyID)
		return err
	}
	return pruneByIDForCompanyUUID(ctx, r.db, "budget_approvals", companyID, ids)
}

func (r *pgBudgetRepo) UpdateBudgetApproval(ctx context.Context, id uuid.UUID, status string, rejectReason *string, resolvedAt time.Time) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE budget_approvals
		SET status = $3, reject_reason = $4, resolved_at = $5
		WHERE company_id = $1 AND id = $2
	`, companyID, id, status, rejectReason, resolvedAt)
	return err
}
