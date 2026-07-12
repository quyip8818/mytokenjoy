package budget

import (
	"context"
	"fmt"
	"sync"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// MonthlyRebalanceScheduler enqueues company-axis rebalance when the open budget month rolls.
type MonthlyRebalanceScheduler struct {
	cfg       config.Config
	store     store.Store
	enqueuer  JobEnqueuer
	lastMonth string
	mu        sync.Mutex
}

func NewMonthlyRebalanceScheduler(cfg config.Config, st store.Store, enqueuer JobEnqueuer) *MonthlyRebalanceScheduler {
	return &MonthlyRebalanceScheduler{
		cfg:       cfg,
		store:     st,
		enqueuer:  enqueuer,
		lastMonth: pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String(),
	}
}

// EnqueueMonthlyRebalanceAll enqueues company-axis rebalance for every active tenant when the month changes.
func (s *MonthlyRebalanceScheduler) EnqueueMonthlyRebalanceAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentMonth := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, s.cfg.Clock()).String()
	if currentMonth == s.lastMonth {
		return nil
	}
	s.lastMonth = currentMonth

	return company.ForEachActiveCompany(ctx, s.store.Company(), func(entryCtx context.Context, co store.Company) error {
		return s.enqueuer.InsertRebalance(entryCtx, co.ID, store.RebalanceAxisCompany, fmt.Sprintf("%d", co.ID))
	})
}
