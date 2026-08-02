package jobs

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

const KindPricingFullSync = "pricing_full_sync"

type PricingFullSyncArgs struct{}

func (PricingFullSyncArgs) Kind() string { return KindPricingFullSync }

func (PricingFullSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueDefault,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}
