package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *billingRepo) GetCurrency(ctx context.Context, code string) (*store.Currency, error) {
	var c store.Currency
	err := r.db.QueryRow(ctx, `
		SELECT currency, points_per_unit, enabled
		FROM currencies
		WHERE currency = $1
	`, code).Scan(&c.Code, &c.PointsPerUnit, &c.Enabled)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
