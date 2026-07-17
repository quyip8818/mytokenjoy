package postgres

import (
	"context"

	"github.com/google/uuid"
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
			id, company_id, amount, currency, points_per_unit, points_granted,
			source, lot_kind, idempotency_key, status,
			display_order_id, payment_method, invoice_status,
			created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, order.ID, order.CompanyID, order.Amount, order.Currency, order.PointsPerUnit, order.PointsGranted,
		order.Source, order.LotKind, order.IdempotencyKey, order.Status,
		order.DisplayOrderID, order.PaymentMethod, order.InvoiceStatus,
		order.CreatedBy, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *billingRepo) GetRechargeOrder(ctx context.Context, id string) (*store.RechargeOrder, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, amount, currency, points_per_unit, points_granted,
			source, lot_kind, idempotency_key, status,
			display_order_id, payment_method, invoice_status,
			created_by, created_at, updated_at
		FROM company_recharge_orders WHERE id = $1
	`, id)
	return scanRechargeOrder(row)
}

func (r *billingRepo) ListRechargeOrders(ctx context.Context, companyID uuid.UUID) ([]store.RechargeOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, amount, currency, points_per_unit, points_granted,
			source, lot_kind, idempotency_key, status,
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
		o, err := scanRechargeOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	return orders, rows.Err()
}

func (r *billingRepo) ConfirmRechargeWithLot(
	ctx context.Context,
	order store.RechargeOrder,
	lot store.RechargeLot,
) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE company_recharge_orders SET
			amount = $2, currency = $3, points_per_unit = $4, points_granted = $5,
			lot_kind = $6, status = $7, updated_at = NOW()
		WHERE id = $1
	`, order.ID, order.Amount, order.Currency, order.PointsPerUnit, order.PointsGranted,
		order.LotKind, order.Status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		if err := r.CreateRechargeOrder(ctx, order); err != nil {
			return err
		}
	}
	lotTag, err := r.db.Exec(ctx, `
		INSERT INTO company_recharge_lots (
			id, company_id, recharge_order_id, billing_currency, lot_kind,
			amount_display, points_granted, points_remaining, unit_price_display,
			status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO NOTHING
	`, lot.ID, lot.CompanyID, lot.RechargeOrderID, lot.BillingCurrency, lot.LotKind,
		lot.AmountDisplay, lot.PointsGranted, lot.PointsRemaining, lot.UnitPriceDisplay,
		lot.Status, lot.CreatedAt, lot.UpdatedAt)
	if err != nil {
		return err
	}
	if lotTag.RowsAffected() == 0 {
		return nil
	}
	return nil
}
