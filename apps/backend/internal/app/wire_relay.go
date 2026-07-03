package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
)

func wirePrecheckService(cfg config.Config, i infra) domainrelay.Prechecker {
	return domainrelay.NewPrecheckService(i.store.Org().Nodes(), i.store.Keys(), i.wallet)
}

func wireGatewayService(cfg config.Config, i infra) (domainrelay.GatewayService, error) {
	return domainrelay.NewGatewayService(cfg, i.store.Relay(), i.store.Company(), wirePrecheckService(cfg, i))
}
