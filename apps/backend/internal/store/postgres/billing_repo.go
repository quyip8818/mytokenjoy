package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type billingRepo struct {
	db dbQuerier
}

func newBillingRepo(db dbQuerier) *billingRepo {
	return &billingRepo{db: db}
}

func (r *billingRepo) CreateRechargeOrder(ctx context.Context, order store.RechargeOrder) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_recharge_orders (
			id, company_id, amount, source, idempotency_key, newapi_topup_ref, status,
			display_order_id, payment_method, invoice_status,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, order.ID, order.CompanyID, order.Amount, order.Source, order.IdempotencyKey,
		order.NewAPITopupRef, order.Status, order.DisplayOrderID, order.PaymentMethod, order.InvoiceStatus,
		order.CreatedBy, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *billingRepo) GetRechargeOrder(ctx context.Context, id string) (*store.RechargeOrder, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, amount, source, idempotency_key, newapi_topup_ref, status,
			display_order_id, payment_method, invoice_status,
			created_by, created_at, updated_at
		FROM company_recharge_orders WHERE id = $1
	`, id)
	var o store.RechargeOrder
	if err := row.Scan(&o.ID, &o.CompanyID, &o.Amount, &o.Source, &o.IdempotencyKey,
		&o.NewAPITopupRef, &o.Status, &o.DisplayOrderID, &o.PaymentMethod, &o.InvoiceStatus,
		&o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *billingRepo) UpdateRechargeStatus(ctx context.Context, id, status string, topupRef *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_recharge_orders SET status = $2, newapi_topup_ref = COALESCE($3, newapi_topup_ref), updated_at = NOW()
		WHERE id = $1
	`, id, status, topupRef)
	return err
}

func (r *billingRepo) ListRechargeOrders(ctx context.Context, companyID int64) ([]store.RechargeOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, amount, source, idempotency_key, newapi_topup_ref, status,
			display_order_id, payment_method, invoice_status,
			created_by, created_at, updated_at
		FROM company_recharge_orders WHERE company_id = $1 ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []store.RechargeOrder
	for rows.Next() {
		var o store.RechargeOrder
		if err := rows.Scan(&o.ID, &o.CompanyID, &o.Amount, &o.Source, &o.IdempotencyKey,
			&o.NewAPITopupRef, &o.Status, &o.DisplayOrderID, &o.PaymentMethod, &o.InvoiceStatus,
			&o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

var _ store.BillingRepository = (*billingRepo)(nil)
