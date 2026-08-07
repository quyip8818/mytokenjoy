package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

// GetCurrency returns the current (latest) row for a currency code.
func (r *billingRepo) GetCurrency(ctx context.Context, code string) (*store.Currency, error) {
	var c store.Currency
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.currency, c.quota_per_unit, c.enabled, c.updated_at, c.updated_by_user_id,
		       COALESCE(u.name, '')
		FROM currencies c
		LEFT JOIN users u ON u.id = c.updated_by_user_id
		WHERE c.currency = $1
		ORDER BY c.updated_at DESC
		LIMIT 1
	`, code).Scan(&c.ID, &c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt, &c.UpdatedByUserID, &c.UpdatedByName)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListEnabledCurrencies returns the latest row per code where enabled=true.
func (r *billingRepo) ListEnabledCurrencies(ctx context.Context) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, currency, quota_per_unit, enabled, updated_at FROM (
			SELECT DISTINCT ON (currency) id, currency, quota_per_unit, enabled, updated_at
			FROM currencies
			ORDER BY currency, updated_at DESC
		) latest
		WHERE enabled = TRUE
		ORDER BY currency
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Currency
	for rows.Next() {
		var c store.Currency
		if err := rows.Scan(&c.ID, &c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListAllCurrencies returns the latest row per code (enabled + disabled).
func (r *billingRepo) ListAllCurrencies(ctx context.Context) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (c.currency) c.id, c.currency, c.quota_per_unit, c.enabled, c.updated_at, c.updated_by_user_id,
		       COALESCE(u.name, '')
		FROM currencies c
		LEFT JOIN users u ON u.id = c.updated_by_user_id
		ORDER BY c.currency, c.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Currency
	for rows.Next() {
		var c store.Currency
		if err := rows.Scan(&c.ID, &c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt, &c.UpdatedByUserID, &c.UpdatedByName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// InsertCurrency appends a new row (append-only). ID is generated if zero.
func (r *billingRepo) InsertCurrency(ctx context.Context, c store.Currency) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.Must(uuid.NewV7())
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO currencies (id, currency, quota_per_unit, enabled, updated_by_user_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, c.ID, c.Code, c.QuotaPerUnit, c.Enabled, c.UpdatedByUserID)
	return err
}

// SetCurrencyEnabled inserts a new row copying current values but with updated enabled status.
func (r *billingRepo) SetCurrencyEnabled(ctx context.Context, code string, enabled bool, actorUserID *uuid.UUID) error {
	// Read current to copy quota_per_unit
	current, err := r.GetCurrency(ctx, code)
	if err != nil {
		return err
	}
	if current == nil {
		return pgx.ErrNoRows
	}
	return r.InsertCurrency(ctx, store.Currency{
		Code:            code,
		QuotaPerUnit:    current.QuotaPerUnit,
		Enabled:         enabled,
		UpdatedByUserID: actorUserID,
	})
}

// ListCurrencyHistory returns all rows for a currency code, newest first.
func (r *billingRepo) ListCurrencyHistory(ctx context.Context, code string, limit, offset int) ([]store.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.currency, c.quota_per_unit, c.enabled, c.updated_at, c.updated_by_user_id,
		       COALESCE(u.name, '')
		FROM currencies c
		LEFT JOIN users u ON u.id = c.updated_by_user_id
		WHERE c.currency = $1
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`, code, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Currency
	for rows.Next() {
		var c store.Currency
		if err := rows.Scan(&c.ID, &c.Code, &c.QuotaPerUnit, &c.Enabled, &c.UpdatedAt, &c.UpdatedByUserID, &c.UpdatedByName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// InsertCurrencyFromSync inserts a currency row from CatalogSync (with known ID). Updates if ID exists.
func (r *billingRepo) InsertCurrencyFromSync(ctx context.Context, c store.Currency) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO currencies (id, currency, quota_per_unit, enabled, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			quota_per_unit = EXCLUDED.quota_per_unit,
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at
	`, c.ID, c.Code, c.QuotaPerUnit, c.Enabled, c.UpdatedAt)
	return err
}

// IsCurrencyReferenced checks if any active company uses this currency code.
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
