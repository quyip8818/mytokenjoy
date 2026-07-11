package newapisync

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

type NewAPISync struct {
	cfg           config.Config
	store         store.Store
	client        newapi.AdminClient
	mappings      store.PlatformKeyMappingRepository
	outbox        store.NewAPISyncOutboxRepository
	wallet        company.WalletService
	channelPolicy ChannelPolicy
}

func New(cfg config.Config, st store.Store, client newapi.AdminClient, wallet company.WalletService, channelPolicy ChannelPolicy) *NewAPISync {
	return &NewAPISync{
		cfg:           cfg,
		store:         st,
		client:        client,
		mappings:      st.PlatformKeyMappings(),
		outbox:        st.AsyncJobs(),
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
