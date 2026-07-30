package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *billingRepo) GetCurrency(ctx context.Context, code string) (*store.Currency, error) {
	var c store.Currency
	err := r.db.QueryRow(ctx, `
		SELECT currency, quota_per_unit, enabled, updated_at
		FROM currencies
		WHERE currency = $1
	`, code).Scan(&c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *billingRepo) ListEnabledCurrencies(ctx context.Context) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT currency, quota_per_unit, enabled, updated_at
		FROM currencies
		WHERE enabled = TRUE
		ORDER BY currency
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCurrencies(rows)
}

func (r *billingRepo) ListAllCurrencies(ctx context.Context) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT currency, quota_per_unit, enabled, updated_at
		FROM currencies
		ORDER BY currency
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCurrencies(rows)
}

func (r *billingRepo) UpsertCurrency(ctx context.Context, c store.Currency) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (currency) DO UPDATE SET
			quota_per_unit = EXCLUDED.quota_per_unit,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()
	`, c.Code, c.QuotaPerUnit, c.Enabled)
	return err
}

func (r *billingRepo) SetCurrencyEnabled(ctx context.Context, code string, enabled bool) error {
	_, err := r.db.Exec(ctx, `
		UPDATE currencies SET enabled = $2, updated_at = NOW()
		WHERE currency = $1
	`, code, enabled)
	return err
}

// ReplaceCurrencies upserts the given currencies and disables any not in the list.
// Should be called within Store.WithTx for atomicity.
func (r *billingRepo) ReplaceCurrencies(ctx context.Context, currencies []store.Currency) error {
	// 1. Upsert all provided currencies
	for _, c := range currencies {
		_, err := r.db.Exec(ctx, `
			INSERT INTO currencies (currency, quota_per_unit, enabled, updated_at)
			VALUES ($1, $2, TRUE, NOW())
			ON CONFLICT (currency) DO UPDATE SET
				quota_per_unit = EXCLUDED.quota_per_unit,
				enabled = TRUE,
				updated_at = NOW()
		`, c.Code, c.QuotaPerUnit)
		if err != nil {
			return err
		}
	}

	// 2. Disable currencies not in the list (FK-safe: no DELETE)
	codes := make([]string, len(currencies))
	for i, c := range currencies {
		codes[i] = c.Code
	}
	_, err := r.db.Exec(ctx, `
		UPDATE currencies SET enabled = FALSE, updated_at = NOW()
		WHERE currency != ALL($1::char(3)[]) AND enabled = TRUE
	`, codes)
	return err
}

func (r *billingRepo) IsCurrencyReferenced(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM companies
			WHERE billing_currency = $1 AND status != 'archived'
		)
	`, code).Scan(&exists)
	return exists, err
}

func scanCurrencies(rows pgx.Rows) ([]store.Currency, error) {
	var out []store.Currency
	for rows.Next() {
		var c store.Currency
		if err := rows.Scan(&c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
