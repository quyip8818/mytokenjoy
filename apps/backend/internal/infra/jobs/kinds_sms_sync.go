package jobs

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type SMSSyncArgs struct{}

func (SMSSyncArgs) Kind() string { return KindSMSSync }

func (SMSSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 7 * 24 * time.Hour,
		},
	}
}
