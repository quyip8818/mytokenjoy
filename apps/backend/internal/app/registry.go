package app

import (
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

// ServiceRegistry holds all wired dependencies for both HTTP and worker layers.
// HTTP layer uses the embedded Deps; worker layer uses Overrun/Rebalance/OrgSync/Infra.
type ServiceRegistry struct {
	httpdeps.Deps
	Infra     infra
	OrgSync   domainorg.SyncService
	Overrun   domainbudget.OverrunProcessor
	Rebalance domainbudget.Rebalancer
}
