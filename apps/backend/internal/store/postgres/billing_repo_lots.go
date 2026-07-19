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
			amount_display, quota_per_unit, quota_granted, quota_remaining,
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
			amount_display, quota_per_unit, quota_granted, quota_remaining,
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
		AmountDisplay: 0, QuotaPerUnit: 1, QuotaGranted: quotaDelta, QuotaRemaining: quotaDelta,
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
