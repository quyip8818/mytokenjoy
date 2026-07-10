package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
)

func wirePrecheckService(cfg config.Config, i infra) domaingateway.Prechecker {
	return domaingateway.NewPrecheckService(
		i.store.BudgetSnapshots(),
		i.store.Org().Nodes(),
		i.store.Budget(),
		i.store.Org(),
		i.store.Keys(),
		i.store.Models(),
		i.wallet,
		i.store.AsyncJobs(),
		cfg.Clock(),
	)
}

func wireGatewayService(cfg config.Config, i infra) (domaingateway.GatewayService, error) {
	return domaingateway.NewGatewayService(cfg, i.store.PlatformKeyMappings(), i.store.Company(), wirePrecheckService(cfg, i))
}
