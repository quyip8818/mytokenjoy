package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgApprovalRepo struct {
	db dbQuerier
}

func (r *pgApprovalRepo) Create(ctx context.Context, req types.ApprovalRequest) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO approval_requests (
			id, company_id, type, status,
			applicant_id, applicant_name, department_id, department_name,
			metadata, approver_id, approver_name, reject_reason,
			created_at, resolved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		req.ID, req.CompanyID, req.Type, req.Status,
		req.ApplicantID, req.ApplicantName, nilUUID(req.DepartmentID), req.DepartmentName,
		req.Metadata, req.ApproverID, req.ApproverName, req.RejectReason,
		req.CreatedAt, req.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("insert approval_request: %w", err)
	}
	return nil
}

func (r *pgApprovalRepo) Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, type, status,
			applicant_id, applicant_name, department_id, department_name,
			metadata, approver_id, approver_name, reject_reason,
			created_at, resolved_at
		FROM approval_requests
		WHERE id = $1 AND company_id = $2
	`, id, companyID)
	return scanApprovalRequest(row)
}

func (r *pgApprovalRepo) Update(ctx context.Context, req types.ApprovalRequest) error {
	_, err := r.db.Exec(ctx, `
		UPDATE approval_requests SET
			status = $3, approver_id = $4, approver_name = $5,
			reject_reason = $6, resolved_at = $7
		WHERE id = $1 AND company_id = $2
	`,
		req.ID, req.CompanyID,
		req.Status, req.ApproverID, req.ApproverName,
		req.RejectReason, req.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("update approval_request: %w", err)
	}
	return nil
}

func (r *pgApprovalRepo) List(ctx context.Context, filter store.ApprovalListFilter) ([]types.ApprovalRequest, int, error) {
	// Build WHERE clause
	args := []any{filter.CompanyID}
	where := "WHERE company_id = $1"
	argIdx := 2

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Type != nil {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.ApplicantID != nil {
		where += fmt.Sprintf(" AND applicant_id = $%d", argIdx)
		args = append(args, *filter.ApplicantID)
		argIdx++
	}

	// Count
	var total int
	countQ := "SELECT count(*) FROM approval_requests " + where
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count approval_requests: %w", err)
	}

	// Fetch
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query := fmt.Sprintf(`
		SELECT id, company_id, type, status,
			applicant_id, applicant_name, department_id, department_name,
			metadata, approver_id, approver_name, reject_reason,
			created_at, resolved_at
		FROM approval_requests %s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, where, limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list approval_requests: %w", err)
	}
	defer rows.Close()

	items := make([]types.ApprovalRequest, 0)
	for rows.Next() {
		item, err := scanApprovalRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// --- scan helpers ---

func scanApprovalRequest(row scannable) (types.ApprovalRequest, error) {
	var req types.ApprovalRequest
	var deptID *uuid.UUID
	var resolvedAt *time.Time
	err := row.Scan(
		&req.ID, &req.CompanyID, &req.Type, &req.Status,
		&req.ApplicantID, &req.ApplicantName, &deptID, &req.DepartmentName,
		&req.Metadata, &req.ApproverID, &req.ApproverName, &req.RejectReason,
		&req.CreatedAt, &resolvedAt,
	)
	if err == pgx.ErrNoRows {
		return types.ApprovalRequest{}, domain.NotFound("approval not found")
	}
	if err != nil {
		return types.ApprovalRequest{}, err
	}
	if deptID != nil {
		req.DepartmentID = *deptID
	}
	req.ResolvedAt = resolvedAt
	return req, nil
}

var _ store.ApprovalRepository = (*pgApprovalRepo)(nil)
