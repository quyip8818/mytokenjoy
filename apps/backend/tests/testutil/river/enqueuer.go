//go:build testhook

package riverfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

// NewInsertOnlyEnqueuer returns an enqueuer backed by a River client that can insert jobs
// without starting workers (for domain tests that only assert enqueue side effects).
func NewInsertOnlyEnqueuer(t *testing.T, cfg config.Config, st store.Store) jobs.Enqueuer {
	t.Helper()
	pool := postgres.MainPool(st)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{DisablePeriodic: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return client.Enqueuer
}

// NewBudgetInsertOnlyEnqueuer wraps NewInsertOnlyEnqueuer as domain/budget.JobEnqueuer.
func NewBudgetInsertOnlyEnqueuer(t *testing.T, cfg config.Config, st store.Store) domainbudget.JobEnqueuer {
	t.Helper()
	return app.NewBudgetEnqueuer(NewInsertOnlyEnqueuer(t, cfg, st))
}

// NewBillingInsertOnlyEnqueuer wraps NewInsertOnlyEnqueuer as domain/billing.JobEnqueuer.
func NewBillingInsertOnlyEnqueuer(t *testing.T, cfg config.Config, st store.Store) domainbilling.JobEnqueuer {
	t.Helper()
	return app.NewBillingEnqueuer(NewInsertOnlyEnqueuer(t, cfg, st))
}

// budgetEnqueuerFromHolder adapts a jobs.Enqueuer for tests using app.BuildRegistry holder.
func budgetEnqueuerFromHolder(enqueuer jobs.Enqueuer) domainbudget.JobEnqueuer {
	return app.NewBudgetEnqueuer(enqueuer)
}
