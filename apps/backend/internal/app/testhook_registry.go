//go:build testhook

package app

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	newapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

func BuildRegistry(cfg config.Config, logger *slog.Logger, st store.Store, opts ...Option) (ServiceRegistry, *jobs.Holder, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	holder := jobs.NewHolder(jobs.NoopEnqueuer{})
	orgAdmin := adapter.NewOrgRiverAdminHolder(nil)
	registry, err := assembleRegistry(cfg, logger, st, o, holder, orgAdmin)
	if err != nil {
		return ServiceRegistry{}, nil, err
	}
	if err := registry.Credentials.BootstrapPlatformIfNeeded(context.Background()); err != nil {
		return ServiceRegistry{}, nil, err
	}
	return registry, holder, nil
}

func (r ServiceRegistry) MustNewAPISync() *newapisync.NewAPISync {
	sync, ok := r.Infra.newAPISync.(*newapisync.NewAPISync)
	if !ok {
		panic("newAPISync is not *newapisync.NewAPISync")
	}
	return sync
}

func (r ServiceRegistry) MustIngestService() *domainusage.IngestService {
	ingest, ok := r.IngestSvc.(*domainusage.IngestService)
	if !ok {
		panic("ingest service is not *domainusage.IngestService")
	}
	return ingest
}
