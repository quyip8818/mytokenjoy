package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *billingRepo) GetCurrency(ctx context.Context, code string) (*store.Currency, error) {
	var c store.Currency
	err := r.db.QueryRow(ctx, `
		SELECT c.currency, c.quota_per_unit, c.enabled, c.updated_at, c.updated_by_user_id,
		       COALESCE(u.name, '')
		FROM currencies c
		LEFT JOIN users u ON u.id = c.updated_by_user_id
		WHERE c.currency = $1
	`, code).Scan(&c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt, &c.UpdatedByUserID, &c.UpdatedByName)
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
	return scanCurrenciesMinimal(rows)
}

func (r *billingRepo) ListAllCurrencies(ctx context.Context) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.currency, c.quota_per_unit, c.enabled, c.updated_at, c.updated_by_user_id,
		       COALESCE(u.name, '')
		FROM currencies c
		LEFT JOIN users u ON u.id = c.updated_by_user_id
		ORDER BY c.currency
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCurrenciesFull(rows)
}

func (r *billingRepo) UpsertCurrency(ctx context.Context, c store.Currency) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled, updated_by_user_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (currency) DO UPDATE SET
			quota_per_unit = EXCLUDED.quota_per_unit,
			enabled = EXCLUDED.enabled,
			updated_by_user_id = EXCLUDED.updated_by_user_id,
			updated_at = NOW()
	`, c.Code, c.QuotaPerUnit, c.Enabled, c.UpdatedByUserID)
	return err
}

func (r *billingRepo) SetCurrencyEnabled(ctx context.Context, code string, enabled bool, actorUserID *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE currencies SET enabled = $2, updated_by_user_id = $3, updated_at = NOW()
		WHERE currency = $1
	`, code, enabled, actorUserID)
	return err
}

// ReplaceCurrencies upserts the given currencies and disables any not in the list.
// CatalogSync only — does NOT touch updated_by_user_id.
// Should be called within Store.WithTx for atomicity.
func (r *billingRepo) ReplaceCurrencies(ctx context.Context, currencies []store.Currency) error {
	// 1. Upsert all provided currencies (not touching updated_by_user_id)
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

// scanCurrenciesMinimal scans rows without JOIN columns (for sync/billing use).
func scanCurrenciesMinimal(rows pgx.Rows) ([]store.Currency, error) {
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

// scanCurrenciesFull scans rows with JOIN columns (for platform admin list).
func scanCurrenciesFull(rows pgx.Rows) ([]store.Currency, error) {
	var out []store.Currency
	for rows.Next() {
		var c store.Currency
		if err := rows.Scan(&c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt, &c.UpdatedByUserID, &c.UpdatedByName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
