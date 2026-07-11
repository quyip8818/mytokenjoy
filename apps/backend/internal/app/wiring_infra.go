package app

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store         store.Store
	adminClient   newapi.AdminClient
	adminPort     adminport.Port
	newAPISync    newapisync.Lifecycle
	channelPolicy newapisync.ChannelPolicy
	wallet        domaincompany.WalletService
	companyGate   *domaincompany.Gate
	notifier      notification.Notifier
	delayer       common.Delayer
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store, adminClientOverride newapi.AdminClient) (infra, error) {
	var adminClient newapi.AdminClient
	if adminClientOverride != nil {
		adminClient = adminClientOverride
	} else {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}
	channelPolicy := newapisync.NewChannelPolicy(cfg)
	wallet := domaincompany.NewWalletService(cfg, adminClient)
	adminPort := newapi.NewAdminPortAdapter(adminClient)

	return infra{
		store:         st,
		adminClient:   adminClient,
		adminPort:     adminPort,
		channelPolicy: channelPolicy,
		wallet:        wallet,
		companyGate:   domaincompany.NewGate(cfg),
		newAPISync:    newapisync.New(cfg, st, adminPort, wallet, channelPolicy),
		notifier:      notification.NewService(cfg, st, logger),
		delayer:       common.NewDelayer(cfg.SimulateDelay),
	}, nil
}
