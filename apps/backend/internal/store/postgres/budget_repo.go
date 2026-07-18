package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

// budgetLockNamespace is used as the first argument to pg_advisory_xact_lock
// to avoid contention with advisory locks from other subsystems.
const budgetLockNamespace = 100

type pgBudgetRepo struct {
	db dbQuerier
}

func (r *pgBudgetRepo) AcquireBudgetLock(ctx context.Context) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1, hashtext($2::text))", budgetLockNamespace, companyID)
	return err
}

func (r *pgBudgetRepo) OrgNodeBudget() store.OrgNodeBudgetRepository {
	return r
}
