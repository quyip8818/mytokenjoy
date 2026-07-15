package budget

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// Async wires budget projection and reconcile with a shared optional gateway cache.
type Async struct {
	Projector *Projector
	Reconcile *ReconcileService
}

// AsyncOption configures optional Async dependencies.
type AsyncOption func(*Async)

// WithProjectorNotifier attaches a notifier for percentage alert checks.
func WithProjectorNotifier(n types.Notifier) AsyncOption {
	return func(a *Async) {
		a.Projector.notifier = n
	}
}

func NewAsync(cfg config.Config, st store.Store, enqueuer JobEnqueuer, cache CombinedKeyCache, logger *slog.Logger, opts ...AsyncOption) *Async {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	if cache == nil {
		cache = NoopCombinedKeyCache
	}
	a := &Async{
		Projector: &Projector{cfg: cfg, store: st, enqueuer: enqueuer, batchSize: defaultProjectorBatchSize, logger: logger, combinedKeyCache: cache},
		Reconcile: &ReconcileService{cfg: cfg, store: st, enqueuer: enqueuer, logger: logger, combinedKeyCache: cache},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}
