package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store       store.Store
	adminClient newapi.AdminClient
	lifecycle   relay.Lifecycle
	notifier    notification.Notifier
	delayer     common.Delayer
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store) (infra, error) {
	var adminClient newapi.AdminClient
	if cfg.NewAPIEnabled {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}

	return infra{
		store:       st,
		adminClient: adminClient,
		lifecycle:   relay.NewTokenLifecycle(cfg, st, adminClient),
		notifier:    notification.NewService(cfg, st, logger),
		delayer:     common.NewDelayer(cfg.SimulateDelay),
	}, nil
}
