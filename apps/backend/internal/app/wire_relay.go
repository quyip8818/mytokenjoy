package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
)

func wirePrecheckService(cfg config.Config, i infra) domainrelay.Prechecker {
	return domainrelay.NewPrecheckService(
		i.store.BudgetSnapshots(),
		i.store.Org().Nodes(),
		i.store.Budget(),
		i.store.Org(),
		i.store.Keys(),
		i.store.Models(),
		i.wallet,
	)
}

func wireGatewayService(cfg config.Config, i infra) (domainrelay.GatewayService, error) {
	return domainrelay.NewGatewayService(cfg, i.store.Relay(), i.store.Company(), wirePrecheckService(cfg, i))
}
