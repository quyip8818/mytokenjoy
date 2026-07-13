package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

func assembleRegistry(cfg config.Config, logger *slog.Logger, st store.Store, o options, holder *jobs.Holder, orgAdmin *OrgRiverAdminHolder) (ServiceRegistry, error) {
	if holder == nil {
		holder = jobs.NewHolder(jobs.NoopEnqueuer{})
	}
	if orgAdmin == nil {
		orgAdmin = NewOrgRiverAdminHolder(nil)
	}
	infraDeps, err := buildInfraWithStore(cfg, logger, st, holder, o.adminClient)
	if err != nil {
		return ServiceRegistry{}, err
	}
	registry := buildServiceRegistry(cfg, infraDeps, buildDomainServices(cfg, infraDeps, logger, holder, orgAdmin))
	if o.orgSync != nil {
		registry.OrgSync = o.orgSync
	}
	return registry, nil
}
