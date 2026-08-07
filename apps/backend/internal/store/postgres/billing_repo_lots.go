package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *billingRepo) ListActiveLotsFIFO(ctx context.Context, companyID uuid.UUID, fifoHeadID *uuid.UUID) ([]store.RechargeLot, error) {
	query := `
		SELECT id, company_id, recharge_order_id, billing_currency, lot_kind,
			paid_amount, quota_per_unit, quota_granted, quota_remaining,
			status, created_at, updated_at
		FROM company_recharge_lots
		WHERE company_id = $1 AND status = $2 AND quota_remaining > 0
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, companyID, store.LotStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lots, err := scanRechargeLots(rows)
	if err != nil {
		return nil, err
	}
	if fifoHeadID == nil || *fifoHeadID == uuid.Nil {
		return lots, nil
	}
	start := 0
	for i, lot := range lots {
		if lot.ID == *fifoHeadID {
			start = i
			break
		}
	}
	if start > 0 {
		lots = append([]store.RechargeLot{}, lots[start:]...)
	}
	return lots, nil
}

func (r *billingRepo) UpdateLotRemaining(ctx context.Context, lot store.RechargeLot) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_recharge_lots
		SET quota_remaining = $2, status = $3, updated_at = NOW()
		WHERE id = $1 AND company_id = $4
	`, lot.ID, lot.QuotaRemaining, lot.Status, lot.CompanyID)
	return err
}

func (r *billingRepo) GetLotByID(ctx context.Context, lotID uuid.UUID) (*store.RechargeLot, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, recharge_order_id, billing_currency, lot_kind,
			paid_amount, quota_per_unit, quota_granted, quota_remaining,
			status, created_at, updated_at
		FROM company_recharge_lots WHERE id = $1
	`, lotID)
	return scanRechargeLot(row)
}

func (r *billingRepo) ExpandOverdraftLot(ctx context.Context, companyID uuid.UUID, billingCurrency string, quotaDelta int64) (*store.RechargeLot, error) {
	key := fmt.Sprintf("overdraft:%s", companyID)
	var existingID uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT l.id FROM company_recharge_lots l
		JOIN company_recharge_orders o ON o.id = l.recharge_order_id
		WHERE l.company_id = $1 AND l.lot_kind = $2 AND l.status = $3
		LIMIT 1
	`, companyID, store.LotKindOverdraft, store.LotStatusActive).Scan(&existingID)
	if err == nil {
		_, err = r.db.Exec(ctx, `
			UPDATE company_recharge_lots
			SET quota_granted = quota_granted + $2,
				quota_remaining = quota_remaining + $2,
				updated_at = NOW()
			WHERE id = $1
		`, existingID, quotaDelta)
		if err != nil {
			return nil, err
		}
		_, err = r.db.Exec(ctx, `
			UPDATE company_recharge_orders
			SET quota_granted = quota_granted + $2, updated_at = NOW()
			WHERE id = (SELECT recharge_order_id FROM company_recharge_lots WHERE id = $1)
		`, existingID, quotaDelta)
		if err != nil {
			return nil, err
		}
		return r.GetLotByID(ctx, existingID)
	}
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	orderID := uuid.Must(uuid.NewV7())
	lotID := orderID
	now := time.Now().UTC()
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: 0, Currency: billingCurrency,
		QuotaPerUnit: 1, QuotaGranted: quotaDelta,
		Source: store.RechargeSourceSystem, LotKind: store.LotKindOverdraft,
		IdempotencyKey: &key, Status: store.RechargeStatusConfirmed,
		CreatedBy: uuid.Nil, CreatedAt: now, UpdatedAt: now,
	}
	lot := store.RechargeLot{
		ID: lotID, CompanyID: companyID, RechargeOrderID: orderID,
		BillingCurrency: billingCurrency, LotKind: store.LotKindOverdraft,
		PaidAmount: 0, QuotaPerUnit: 1, QuotaGranted: quotaDelta, QuotaRemaining: quotaDelta,
		Status: store.LotStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := r.ConfirmRechargeWithLot(ctx, order, lot); err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *billingRepo) ExpireMockLots(ctx context.Context, companyID uuid.UUID) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE company_recharge_lots
		SET status = 'expired', updated_at = NOW()
		WHERE company_id = $1 AND lot_kind = $2 AND status = $3
	`, companyID, store.LotKindMock, store.LotStatusActive)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *billingRepo) SumActiveLotsRemaining(ctx context.Context, companyID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(quota_remaining), 0)
		FROM company_recharge_lots
		WHERE company_id = $1 AND status = $2
	`, companyID, store.LotStatusActive).Scan(&total)
	return total, err
}

// UpsertOrderFromSync inserts a recharge order from platform sync or updates if it already exists.
func (r *billingRepo) UpsertOrderFromSync(ctx context.Context, order store.RechargeOrder) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_recharge_orders (id, company_id, amount, currency, quota_per_unit, quota_granted, source, lot_kind, status, display_order_id, payment_method, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
		ON CONFLICT (id) DO UPDATE SET
			amount = EXCLUDED.amount,
			status = EXCLUDED.status,
			quota_granted = EXCLUDED.quota_granted,
			updated_at = NOW()
	`, order.ID, order.CompanyID, order.Amount, order.Currency,
		order.QuotaPerUnit, order.QuotaGranted, order.Source, order.LotKind,
		order.Status, order.DisplayOrderID, order.PaymentMethod, uuid.Nil, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert order from sync: %w", err)
	}
	return nil
}

// UpsertLotFromSync inserts a lot from platform sync or updates quota_remaining/status if it already exists.
func (r *billingRepo) UpsertLotFromSync(ctx context.Context, lot store.RechargeLot) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_recharge_lots (id, company_id, recharge_order_id, billing_currency, lot_kind, paid_amount, quota_per_unit, quota_granted, quota_remaining, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (id) DO UPDATE SET
			quota_remaining = EXCLUDED.quota_remaining,
			status = EXCLUDED.status,
			updated_at = NOW()
	`, lot.ID, lot.CompanyID, lot.RechargeOrderID, lot.BillingCurrency, lot.LotKind,
		lot.PaidAmount, lot.QuotaPerUnit, lot.QuotaGranted, lot.QuotaRemaining, lot.Status, lot.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert lot: %w", err)
	}
	return nil
}
