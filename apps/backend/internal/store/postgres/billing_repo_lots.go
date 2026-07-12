package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *billingRepo) ListActiveLotsFIFO(ctx context.Context, companyID int64, fifoHeadID *string) ([]store.RechargeLot, error) {
	query := `
		SELECT id, company_id, recharge_order_id, billing_currency, lot_kind,
			amount_display, points_granted, points_remaining, unit_price_display,
			status, created_at, updated_at
		FROM company_recharge_lots
		WHERE company_id = $1 AND status = $2 AND points_remaining > 0
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
	if fifoHeadID == nil || *fifoHeadID == "" {
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
		SET points_remaining = $2, status = $3, updated_at = NOW()
		WHERE id = $1 AND company_id = $4
	`, lot.ID, lot.PointsRemaining, lot.Status, lot.CompanyID)
	return err
}

func (r *billingRepo) GetLotByID(ctx context.Context, lotID string) (*store.RechargeLot, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, recharge_order_id, billing_currency, lot_kind,
			amount_display, points_granted, points_remaining, unit_price_display,
			status, created_at, updated_at
		FROM company_recharge_lots WHERE id = $1
	`, lotID)
	return scanRechargeLot(row)
}

func (r *billingRepo) ExpandOverdraftLot(ctx context.Context, companyID int64, billingCurrency string, pointsDelta float64) (*store.RechargeLot, error) {
	key := fmt.Sprintf("overdraft:%d", companyID)
	var existingID string
	err := r.db.QueryRow(ctx, `
		SELECT l.id FROM company_recharge_lots l
		JOIN company_recharge_orders o ON o.id = l.recharge_order_id
		WHERE l.company_id = $1 AND l.lot_kind = $2 AND l.status = $3
		LIMIT 1
	`, companyID, store.LotKindOverdraft, store.LotStatusActive).Scan(&existingID)
	if err == nil {
		_, err = r.db.Exec(ctx, `
			UPDATE company_recharge_lots
			SET points_granted = points_granted + $2,
				points_remaining = points_remaining + $2,
				updated_at = NOW()
			WHERE id = $1
		`, existingID, pointsDelta)
		if err != nil {
			return nil, err
		}
		_, err = r.db.Exec(ctx, `
			UPDATE company_recharge_orders
			SET points_granted = points_granted + $2, updated_at = NOW()
			WHERE id = (SELECT recharge_order_id FROM company_recharge_lots WHERE id = $1)
		`, existingID, pointsDelta)
		if err != nil {
			return nil, err
		}
		return r.GetLotByID(ctx, existingID)
	}
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	orderID := fmt.Sprintf("od-%d", companyID)
	lotID := orderID
	now := time.Now().UTC()
	order := store.RechargeOrder{
		ID: orderID, CompanyID: companyID, Amount: 0, Currency: billingCurrency,
		PointsPerUnit: 1, PointsGranted: pointsDelta,
		Source: store.RechargeSourceSystem, LotKind: store.LotKindOverdraft,
		IdempotencyKey: &key, Status: store.RechargeStatusConfirmed,
		CreatedBy: "system", CreatedAt: now, UpdatedAt: now,
	}
	lot := store.RechargeLot{
		ID: lotID, CompanyID: companyID, RechargeOrderID: orderID,
		BillingCurrency: billingCurrency, LotKind: store.LotKindOverdraft,
		AmountDisplay: 0, PointsGranted: pointsDelta, PointsRemaining: pointsDelta,
		UnitPriceDisplay: 0, Status: store.LotStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := r.ConfirmRechargeWithLot(ctx, order, lot); err != nil {
		return nil, err
	}
	return &lot, nil
}
