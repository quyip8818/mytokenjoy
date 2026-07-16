package syncdeps

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/store"
)

type Deps struct {
	Cfg           config.Config
	Store         store.Store
	Client        adminport.Port
	Mappings      store.PlatformKeyMappingRepository
	Enqueuer      ports.SyncJobEnqueuer
	ChannelPolicy policy.ChannelPolicy
}

func Enabled(d Deps) bool {
	return d.Client != nil
}
