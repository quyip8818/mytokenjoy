package newapisync

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type NewAPISync struct {
	cfg           config.Config
	store         store.Store
	client        adminport.Port
	mappings      store.PlatformKeyMappingRepository
	enqueuer      jobs.Enqueuer
	wallet        company.WalletService
	channelPolicy ChannelPolicy
}

func New(cfg config.Config, st store.Store, client adminport.Port, wallet company.WalletService, channelPolicy ChannelPolicy, enqueuer jobs.Enqueuer) *NewAPISync {
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
	}
	return &NewAPISync{
		cfg:           cfg,
		store:         st,
		client:        client,
		mappings:      st.PlatformKeyMappings(),
		enqueuer:      enqueuer,
		wallet:        wallet,
		channelPolicy: channelPolicy,
	}
}

var (
	_ Lifecycle            = (*NewAPISync)(nil)
	_ ModelLimitsLifecycle = (*NewAPISync)(nil)
	_ OverrunKeyControl    = (*NewAPISync)(nil)
	_ KeysNewAPISync       = (*NewAPISync)(nil)
	_ OutboxHandler        = (*NewAPISync)(nil)
)

func (l *NewAPISync) Enabled() bool {
	return l.client != nil
}
