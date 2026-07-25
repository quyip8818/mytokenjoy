package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sms/backend/internal/domain/order"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListOrders(ctx context.Context, f order.ListFilter) (*types.PagedResult[types.PurchaseOrder], error) {
	var conditions []string
	var args []any
	idx := 1

	if f.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("o.order_no ILIKE $%d", idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	if f.SupplierID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("o.supplier_id = $%d", idx))
		args = append(args, f.SupplierID)
		idx++
	}
	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM purchase_orders o "+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT o.id, o.order_no, o.supplier_id, o.contract_id, o.total_amount, o.order_date, o.status,
		        o.description, o.created_by, o.created_at, o.updated_at,
		        sup.name as supplier_name, c.contract_no, u.real_name as creator_name
		 FROM purchase_orders o
		 LEFT JOIN suppliers sup ON sup.id = o.supplier_id
		 LEFT JOIN contracts c ON c.id = o.contract_id
		 LEFT JOIN users u ON u.id = o.created_by
		 %s ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.PurchaseOrder
	for rows.Next() {
		var o types.PurchaseOrder
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.SupplierID, &o.ContractID, &o.TotalAmount, &o.OrderDate,
			&o.Status, &o.Description, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
			&o.SupplierName, &o.ContractNo, &o.CreatorName); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	if items == nil {
		items = []types.PurchaseOrder{}
	}
	return &types.PagedResult[types.PurchaseOrder]{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Store) GetOrder(ctx context.Context, id uuid.UUID) (*types.PurchaseOrder, error) {
	var o types.PurchaseOrder
	err := s.pool.QueryRow(ctx,
		`SELECT o.id, o.order_no, o.supplier_id, o.contract_id, o.total_amount, o.order_date, o.status,
		        o.description, o.created_by, o.created_at, o.updated_at,
		        sup.name, c.contract_no, u.real_name
		 FROM purchase_orders o
		 LEFT JOIN suppliers sup ON sup.id = o.supplier_id
		 LEFT JOIN contracts c ON c.id = o.contract_id
		 LEFT JOIN users u ON u.id = o.created_by
		 WHERE o.id = $1`, id,
	).Scan(&o.ID, &o.OrderNo, &o.SupplierID, &o.ContractID, &o.TotalAmount, &o.OrderDate,
		&o.Status, &o.Description, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
		&o.SupplierName, &o.ContractNo, &o.CreatorName)
	return &o, store.WrapNotFound(err)
}

func (s *Store) CreateOrder(ctx context.Context, o *types.PurchaseOrder) error {
	o.ID = newID()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO purchase_orders (id, order_no, supplier_id, contract_id, total_amount, order_date, status, description, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING created_at, updated_at`,
		o.ID, o.OrderNo, o.SupplierID, o.ContractID, o.TotalAmount, o.OrderDate, o.Status, o.Description, o.CreatedBy,
	).Scan(&o.CreatedAt, &o.UpdatedAt)
	return store.WrapConflict(err)
}

func (s *Store) UpdateOrder(ctx context.Context, id uuid.UUID, o *types.PurchaseOrder) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE purchase_orders SET order_no=$1, supplier_id=$2, contract_id=$3,
		 total_amount=$4, order_date=$5, status=$6, description=$7 WHERE id=$8`,
		o.OrderNo, o.SupplierID, o.ContractID, o.TotalAmount, o.OrderDate, o.Status, o.Description, id,
	)
	if err != nil {
		return store.WrapConflict(err)
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM purchase_orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) RecentOrders(ctx context.Context, limit int) ([]types.PurchaseOrder, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT o.id, o.order_no, o.supplier_id, o.contract_id, o.total_amount, o.order_date, o.status,
		        o.description, o.created_by, o.created_at, o.updated_at,
		        sup.name, c.contract_no, u.real_name
		 FROM purchase_orders o
		 LEFT JOIN suppliers sup ON sup.id = o.supplier_id
		 LEFT JOIN contracts c ON c.id = o.contract_id
		 LEFT JOIN users u ON u.id = o.created_by
		 ORDER BY o.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.PurchaseOrder
	for rows.Next() {
		var o types.PurchaseOrder
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.SupplierID, &o.ContractID, &o.TotalAmount, &o.OrderDate,
			&o.Status, &o.Description, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
			&o.SupplierName, &o.ContractNo, &o.CreatorName); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	if items == nil {
		items = []types.PurchaseOrder{}
	}
	return items, nil
}
