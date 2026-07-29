package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sms/backend/internal/domain/supplier"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListSuppliers(ctx context.Context, f supplier.ListFilter) (*types.PagedResult[types.Supplier], error) {
	var conditions []string
	var args []any
	idx := 1

	if f.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", idx))
		args = append(args, f.Category)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM suppliers "+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT id, name, code, category, website, status, description, created_by, created_at, updated_at
		 FROM suppliers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.Supplier
	for rows.Next() {
		var sup types.Supplier
		if err := rows.Scan(&sup.ID, &sup.Name, &sup.Code, &sup.Category, &sup.Website,
			&sup.Status, &sup.Description, &sup.CreatedBy, &sup.CreatedAt, &sup.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, sup)
	}
	if items == nil {
		items = []types.Supplier{}
	}
	return &types.PagedResult[types.Supplier]{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Store) GetSupplier(ctx context.Context, id uuid.UUID) (*types.SupplierDetail, error) {
	var sup types.Supplier
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, code, category, website, status, description, created_by, created_at, updated_at
		 FROM suppliers WHERE id = $1`, id,
	).Scan(&sup.ID, &sup.Name, &sup.Code, &sup.Category, &sup.Website,
		&sup.Status, &sup.Description, &sup.CreatedBy, &sup.CreatedAt, &sup.UpdatedAt)
	if err != nil {
		return nil, store.WrapNotFound(err)
	}

	// Contacts
	contactRows, err := s.pool.Query(ctx,
		`SELECT id, supplier_id, name, position, phone, email, is_primary, created_at
		 FROM supplier_contacts WHERE supplier_id = $1 ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer contactRows.Close()
	var contacts []types.SupplierContact
	for contactRows.Next() {
		var c types.SupplierContact
		if err := contactRows.Scan(&c.ID, &c.SupplierID, &c.Name, &c.Position, &c.Phone, &c.Email, &c.IsPrimary, &c.CreatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	if contacts == nil {
		contacts = []types.SupplierContact{}
	}

	// Contracts
	contractListRows, err := s.pool.Query(ctx,
		`SELECT id, supplier_id, contract_no, title, amount, sign_date, start_date, end_date, status, remarks, created_by, created_at, updated_at
		 FROM contracts WHERE supplier_id = $1 ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer contractListRows.Close()
	var contracts []types.Contract
	for contractListRows.Next() {
		var c types.Contract
		if err := contractListRows.Scan(&c.ID, &c.SupplierID, &c.ContractNo, &c.Title, &c.Amount,
			&c.SignDate, &c.StartDate, &c.EndDate, &c.Status, &c.Remarks, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	if contracts == nil {
		contracts = []types.Contract{}
	}

	// Orders
	orderRows, err := s.pool.Query(ctx,
		`SELECT o.id, o.order_no, o.supplier_id, o.contract_id, o.total_amount, o.order_date, o.status,
		        o.description, o.created_by, o.created_at, o.updated_at,
		        c.contract_no
		 FROM purchase_orders o LEFT JOIN contracts c ON c.id = o.contract_id
		 WHERE o.supplier_id = $1 ORDER BY o.created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer orderRows.Close()
	var orders []types.PurchaseOrder
	for orderRows.Next() {
		var o types.PurchaseOrder
		if err := orderRows.Scan(&o.ID, &o.OrderNo, &o.SupplierID, &o.ContractID, &o.TotalAmount, &o.OrderDate,
			&o.Status, &o.Description, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt, &o.ContractNo); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []types.PurchaseOrder{}
	}

	// Evaluations
	evalRows, err := s.pool.Query(ctx,
		`SELECT e.id, e.supplier_id, e.evaluator_id, e.period, e.quality, e.performance, e.price, e.service, e.compliance,
		        e.total_score, e.grade, e.comment, e.created_at, u.real_name as evaluator_name
		 FROM evaluations e LEFT JOIN users u ON u.id = e.evaluator_id
		 WHERE e.supplier_id = $1 ORDER BY e.created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer evalRows.Close()
	var evaluations []types.Evaluation
	for evalRows.Next() {
		var ev types.Evaluation
		if err := evalRows.Scan(&ev.ID, &ev.SupplierID, &ev.EvaluatorID, &ev.Period,
			&ev.Quality, &ev.Performance, &ev.Price, &ev.Service, &ev.Compliance,
			&ev.TotalScore, &ev.Grade, &ev.Comment, &ev.CreatedAt, &ev.EvaluatorName); err != nil {
			return nil, err
		}
		evaluations = append(evaluations, ev)
	}
	if evaluations == nil {
		evaluations = []types.Evaluation{}
	}

	return &types.SupplierDetail{
		Supplier:    sup,
		Contacts:    contacts,
		Contracts:   contracts,
		Orders:      orders,
		Evaluations: evaluations,
	}, nil
}

func (s *Store) CreateSupplier(ctx context.Context, sup *types.Supplier) error {
	sup.ID = newID()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO suppliers (id, name, code, category, website, status, description, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING created_at, updated_at`,
		sup.ID, sup.Name, sup.Code, sup.Category, sup.Website, sup.Status, sup.Description, sup.CreatedBy,
	).Scan(&sup.CreatedAt, &sup.UpdatedAt)
	return store.WrapConflict(err)
}

func (s *Store) UpdateSupplier(ctx context.Context, id uuid.UUID, sup *types.Supplier) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE suppliers SET name=$1, code=$2, category=$3, website=$4, status=$5, description=$6
		 WHERE id=$7`,
		sup.Name, sup.Code, sup.Category, sup.Website, sup.Status, sup.Description, id,
	)
	if err != nil {
		return store.WrapConflict(err)
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteSupplier(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM suppliers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) HasSupplierRefs(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM contracts WHERE supplier_id=$1) +
		        (SELECT count(*) FROM purchase_orders WHERE supplier_id=$1)`, id,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) SupplierOptions(ctx context.Context) ([]types.IDName, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name FROM suppliers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.IDName
	for rows.Next() {
		var item types.IDName
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []types.IDName{}
	}
	return items, nil
}

func (s *Store) CreateContact(ctx context.Context, c *types.SupplierContact) error {
	c.ID = newID()
	return s.pool.QueryRow(ctx,
		`INSERT INTO supplier_contacts (id, supplier_id, name, position, phone, email, is_primary)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at`,
		c.ID, c.SupplierID, c.Name, c.Position, c.Phone, c.Email, c.IsPrimary,
	).Scan(&c.CreatedAt)
}

func (s *Store) UpdateContact(ctx context.Context, c *types.SupplierContact) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE supplier_contacts SET name=$1, position=$2, phone=$3, email=$4, is_primary=$5 WHERE id=$6`,
		c.Name, c.Position, c.Phone, c.Email, c.IsPrimary, c.ID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteContact(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM supplier_contacts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}
