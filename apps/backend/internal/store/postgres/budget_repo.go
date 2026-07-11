package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type pgBudgetRepo struct {
	db dbQuerier
}

func (r *pgBudgetRepo) AcquireBudgetLock(ctx context.Context) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", companyID)
	return err
}
