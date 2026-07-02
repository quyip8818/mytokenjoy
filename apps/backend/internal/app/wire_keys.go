package app

import (
	"github.com/tokenjoy/backend/internal/config"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
)

func wireKeys(cfg config.Config, i infra) domainkeys.Service {
	return domainkeys.NewService(cfg, i.store, i.lifecycle, i.delayer)
}

func wireModels(cfg config.Config, i infra) domainmodels.Service {
	return domainmodels.NewService(cfg, i.store, i.adminClient, i.lifecycle, i.delayer)
}
