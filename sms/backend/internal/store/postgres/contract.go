package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sms/backend/internal/domain/contract"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListContracts(ctx context.Context, f contract.ListFilter) (*types.PagedResult[types.Contract], error) {
	var conditions []string
	var args []any
	idx := 1

	if f.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(c.title ILIKE $%d OR c.contract_no ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	if f.SupplierID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("c.supplier_id = $%d", idx))
		args = append(args, f.SupplierID)
		idx++
	}
	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("c.status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM contracts c "+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT c.id, c.supplier_id, c.contract_no, c.title, c.amount, c.sign_date, c.start_date, c.end_date,
		        c.status, c.remarks, c.created_by, c.created_at, c.updated_at, sup.name as supplier_name
		 FROM contracts c LEFT JOIN suppliers sup ON sup.id = c.supplier_id
		 %s ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.Contract
	for rows.Next() {
		var c types.Contract
		if err := rows.Scan(&c.ID, &c.SupplierID, &c.ContractNo, &c.Title, &c.Amount,
			&c.SignDate, &c.StartDate, &c.EndDate, &c.Status, &c.Remarks, &c.CreatedBy,
			&c.CreatedAt, &c.UpdatedAt, &c.SupplierName); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []types.Contract{}
	}
	return &types.PagedResult[types.Contract]{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Store) GetContract(ctx context.Context, id uuid.UUID) (*types.ContractDetail, error) {
	var c types.Contract
	err := s.pool.QueryRow(ctx,
		`SELECT c.id, c.supplier_id, c.contract_no, c.title, c.amount, c.sign_date, c.start_date, c.end_date,
		        c.status, c.remarks, c.created_by, c.created_at, c.updated_at, sup.name as supplier_name
		 FROM contracts c LEFT JOIN suppliers sup ON sup.id = c.supplier_id
		 WHERE c.id = $1`, id,
	).Scan(&c.ID, &c.SupplierID, &c.ContractNo, &c.Title, &c.Amount,
		&c.SignDate, &c.StartDate, &c.EndDate, &c.Status, &c.Remarks, &c.CreatedBy,
		&c.CreatedAt, &c.UpdatedAt, &c.SupplierName)
	if err != nil {
		return nil, store.WrapNotFound(err)
	}

	atts, err := s.ListAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.ContractDetail{Contract: c, Attachments: atts}, nil
}

func (s *Store) CreateContract(ctx context.Context, c *types.Contract) error {
	c.ID = newID()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO contracts (id, supplier_id, contract_no, title, amount, sign_date, start_date, end_date, status, remarks, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING created_at, updated_at`,
		c.ID, c.SupplierID, c.ContractNo, c.Title, c.Amount, c.SignDate, c.StartDate, c.EndDate, c.Status, c.Remarks, c.CreatedBy,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	return store.WrapConflict(err)
}

func (s *Store) UpdateContract(ctx context.Context, id uuid.UUID, c *types.Contract) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE contracts SET supplier_id=$1, contract_no=$2, title=$3, amount=$4,
		 sign_date=$5, start_date=$6, end_date=$7, status=$8, remarks=$9 WHERE id=$10`,
		c.SupplierID, c.ContractNo, c.Title, c.Amount, c.SignDate, c.StartDate, c.EndDate, c.Status, c.Remarks, id,
	)
	if err != nil {
		return store.WrapConflict(err)
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteContract(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) HasContractOrders(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM purchase_orders WHERE contract_id = $1`, id).Scan(&count)
	return count > 0, err
}

func (s *Store) CreateAttachment(ctx context.Context, a *types.ContractAttachment) error {
	a.ID = newID()
	return s.pool.QueryRow(ctx,
		`INSERT INTO contract_attachments (id, contract_id, file_name, file_path, file_size, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`,
		a.ID, a.ContractID, a.FileName, a.FilePath, a.FileSize, a.UploadedBy,
	).Scan(&a.CreatedAt)
}

func (s *Store) GetAttachment(ctx context.Context, id uuid.UUID) (*types.ContractAttachment, error) {
	var a types.ContractAttachment
	err := s.pool.QueryRow(ctx,
		`SELECT id, contract_id, file_name, file_path, file_size, uploaded_by, created_at
		 FROM contract_attachments WHERE id = $1`, id,
	).Scan(&a.ID, &a.ContractID, &a.FileName, &a.FilePath, &a.FileSize, &a.UploadedBy, &a.CreatedAt)
	return &a, store.WrapNotFound(err)
}

func (s *Store) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM contract_attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) ListAttachments(ctx context.Context, contractID uuid.UUID) ([]types.ContractAttachment, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, contract_id, file_name, file_path, file_size, uploaded_by, created_at
		 FROM contract_attachments WHERE contract_id = $1 ORDER BY created_at`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.ContractAttachment
	for rows.Next() {
		var a types.ContractAttachment
		if err := rows.Scan(&a.ID, &a.ContractID, &a.FileName, &a.FilePath, &a.FileSize, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	if items == nil {
		items = []types.ContractAttachment{}
	}
	return items, nil
}
