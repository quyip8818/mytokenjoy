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
	// ponytail: hashtext produces 32-bit int from UUID text, so two different companyIDs
	// could theoretically collide on the same advisory lock (~1 in 4B).
	// Acceptable for budget serialization; upgrade path: use the low 64 bits of UUID directly.
	_, err := r.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1, hashtext($2::text))", budgetLockNamespace, companyID)
	return err
}

func (r *pgBudgetRepo) OrgNodeBudget() store.OrgNodeBudgetRepository {
	return r
}
