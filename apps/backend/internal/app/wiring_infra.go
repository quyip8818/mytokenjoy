package app

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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
	enqueuer      jobs.Enqueuer
	budgetCheck   budgetcheck.Store
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store, enqueuer jobs.Enqueuer, adminClientOverride newapi.AdminClient) (infra, error) {
	var adminClient newapi.AdminClient
	if adminClientOverride != nil {
		adminClient = adminClientOverride
	} else {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
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
		newAPISync:    newapisync.New(cfg, st, adminPort, wallet, channelPolicy, enqueuer),
		notifier:      notification.NewService(cfg, st, logger),
		delayer:       common.NewDelayer(cfg.SimulateDelay),
		enqueuer:      enqueuer,
		budgetCheck:   budgetcheck.Open(context.Background(), cfg, logger),
	}, nil
}
