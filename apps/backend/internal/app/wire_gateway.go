package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
)

func wirePrecheckService(cfg config.Config, i infra) domaingateway.Prechecker {
	return domaingateway.NewPrecheckService(i.store.GatewayPrecheck(), cfg.Clock(), i.budgetCheck)
}

func wireGatewayService(cfg config.Config, i infra) (domaingateway.GatewayService, error) {
	return domaingateway.NewGatewayService(cfg, wirePrecheckService(cfg, i))
}
