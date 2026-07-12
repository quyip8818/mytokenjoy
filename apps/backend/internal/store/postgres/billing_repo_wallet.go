package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

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
