package jobs

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type CatalogSyncArgs struct{}

func (CatalogSyncArgs) Kind() string { return KindCatalogSync }

func (CatalogSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueDefault,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}
