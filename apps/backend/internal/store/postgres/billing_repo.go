package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

func (r *billingRepo) ListRechargeOrders(ctx context.Context, companyID int64) ([]store.RechargeOrder, error) {
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
		WHERE id = $1
	`, lot.ID, lot.PointsRemaining, lot.Status)
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

func (r *billingRepo) AggregateWallet(ctx context.Context, companyID int64) (store.WalletAggregate, error) {
	var billingCurrency string
	var walletRemain float64
	if err := r.db.QueryRow(ctx, `SELECT billing_currency, wallet_remain FROM companies WHERE id = $1`, companyID).
		Scan(&billingCurrency, &walletRemain); err != nil {
		return store.WalletAggregate{}, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT billing_currency,
			COALESCE(SUM(CASE WHEN lot_kind IN ('paid','adjust') THEN amount_display ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lot_kind IN ('paid','adjust') THEN points_remaining * unit_price_display ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lot_kind = 'gift' THEN points_remaining ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lot_kind = 'overdraft' THEN points_remaining ELSE 0 END), 0)
		FROM company_recharge_lots
		WHERE company_id = $1
		GROUP BY billing_currency
	`, companyID)
	if err != nil {
		return store.WalletAggregate{}, err
	}
	defer rows.Close()
	var balances []store.WalletCurrencyBalance
	var giftPoints, overdraftPoints float64
	for rows.Next() {
		var currency string
		var topup, balance, gift, overdraft float64
		if err := rows.Scan(&currency, &topup, &balance, &gift, &overdraft); err != nil {
			return store.WalletAggregate{}, err
		}
		balances = append(balances, store.WalletCurrencyBalance{
			Currency:      currency,
			TotalTopup:    topup,
			Balance:       balance,
			TotalConsumed: topup - balance,
		})
		giftPoints += gift
		overdraftPoints += overdraft
	}
	return store.WalletAggregate{
		BillingCurrency: billingCurrency,
		Balances:        balances,
		WalletRemain:    walletRemain,
		GiftPoints:      giftPoints,
		OverdraftPoints: overdraftPoints,
	}, rows.Err()
}

type rechargeScanner interface {
	Scan(dest ...any) error
}

func scanRechargeOrder(row rechargeScanner) (*store.RechargeOrder, error) {
	var o store.RechargeOrder
	if err := row.Scan(
		&o.ID, &o.CompanyID, &o.Amount, &o.Currency, &o.PointsPerUnit, &o.PointsGranted,
		&o.Source, &o.LotKind, &o.IdempotencyKey, &o.Status,
		&o.DisplayOrderID, &o.PaymentMethod, &o.InvoiceStatus,
		&o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &o, nil
}

func scanRechargeLot(row rechargeScanner) (*store.RechargeLot, error) {
	var lot store.RechargeLot
	if err := row.Scan(
		&lot.ID, &lot.CompanyID, &lot.RechargeOrderID, &lot.BillingCurrency, &lot.LotKind,
		&lot.AmountDisplay, &lot.PointsGranted, &lot.PointsRemaining, &lot.UnitPriceDisplay,
		&lot.Status, &lot.CreatedAt, &lot.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &lot, nil
}

func scanRechargeLots(rows pgx.Rows) ([]store.RechargeLot, error) {
	var lots []store.RechargeLot
	for rows.Next() {
		lot, err := scanRechargeLot(rows)
		if err != nil {
			return nil, err
		}
		lots = append(lots, *lot)
	}
	return lots, rows.Err()
}

var _ store.BillingRepository = (*billingRepo)(nil)
