package budget

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

// Async wires budget projection and reconcile with a shared optional gateway cache.
type Async struct {
	Projector *Projector
	Reconcile *ReconcileService
}

func NewAsync(cfg config.Config, st store.Store, enqueuer jobs.Enqueuer, cache budgetcheck.Store, logger *slog.Logger) *Async {
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
	}
	if cache == nil {
		cache = budgetcheck.Noop{}
	}
	return &Async{
		Projector: &Projector{cfg: cfg, store: st, enqueuer: enqueuer, batchSize: defaultProjectorBatchSize, logger: logger, gatewayCache: cache},
		Reconcile: &ReconcileService{cfg: cfg, store: st, enqueuer: enqueuer, logger: logger, gatewayCache: cache},
	}
}
