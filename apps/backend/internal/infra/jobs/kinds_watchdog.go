package jobs

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type TenantWatchdogArgs struct{}

func (TenantWatchdogArgs) Kind() string { return KindTenantWatchdog }

func (TenantWatchdogArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 7 * 24 * time.Hour,
		},
	}
}
