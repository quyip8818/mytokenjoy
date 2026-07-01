package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store         store.Store
	adminClient   newapi.AdminClient
	lifecycle     relay.Lifecycle
	channelPolicy relay.ChannelPolicy
	wallet        domaincompany.WalletService
	companyGate   *domaincompany.Gate
	notifier      notification.Notifier
	delayer       common.Delayer
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store) (infra, error) {
	var adminClient newapi.AdminClient
	if cfg.NewAPIEnabled {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}
	channelPolicy := relay.NewChannelPolicy(cfg)
	wallet := domaincompany.NewWalletService(cfg, adminClient)

	return infra{
		store:         st,
		adminClient:   adminClient,
		channelPolicy: channelPolicy,
		wallet:        wallet,
		companyGate:   domaincompany.NewGate(cfg),
		lifecycle:     relay.NewTokenLifecycle(cfg, st, adminClient, wallet, channelPolicy),
		notifier:      notification.NewService(cfg, st, logger),
		delayer:       common.NewDelayer(cfg.SimulateDelay),
	}, nil
}
