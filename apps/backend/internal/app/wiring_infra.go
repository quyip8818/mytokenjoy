package app

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store         store.Store
	adminPort     adminport.Port
	newAPISync    newapisync.Lifecycle
	channelPolicy newapisync.ChannelPolicy
	wallet        domaincompany.WalletService
	companyGate   *domaincompany.Gate
	notifier      types.Notifier
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
	adminPort := newapi.NewAdminPortAdapter(adminClient)
	wallet := domaincompany.NewWalletService(cfg, adminPort)

	return infra{
		store:         st,
		adminPort:     adminPort,
		channelPolicy: channelPolicy,
		wallet:        wallet,
		companyGate:   domaincompany.NewGate(cfg),
		newAPISync:    newapisync.New(cfg, st, adminPort, wallet, channelPolicy, NewNewAPISyncEnqueuer(enqueuer)),
		notifier:      notification.NewService(cfg, st, logger),
		delayer:       common.NewDelayer(cfg.SimulateDelay),
		enqueuer:      enqueuer,
		budgetCheck:   budgetcheck.Open(context.Background(), cfg, logger),
	}, nil
}
