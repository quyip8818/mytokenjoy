package relay

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

type TokenLifecycle struct {
	cfg           config.Config
	store         store.Store
	client        newapi.AdminClient
	mappings      store.RelayMappingRepository
	relayOutbox   store.RelayOutboxRepository
	wallet        company.WalletService
	channelPolicy ChannelPolicy
}

func NewTokenLifecycle(cfg config.Config, st store.Store, client newapi.AdminClient, wallet company.WalletService, channelPolicy ChannelPolicy) *TokenLifecycle {
	relayRepo := st.Relay()
	return &TokenLifecycle{
		cfg:           cfg,
		store:         st,
		client:        client,
		mappings:      relayRepo,
		relayOutbox:   relayRepo,
		wallet:        wallet,
		channelPolicy: channelPolicy,
	}
}

var _ Lifecycle = (*TokenLifecycle)(nil)

func (l *TokenLifecycle) Enabled() bool {
	return l.cfg.NewAPIEnabled && l.client != nil
}
