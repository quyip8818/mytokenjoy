package jobs

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type SMSSyncArgs struct{}

func (SMSSyncArgs) Kind() string { return KindSMSSync }

func (SMSSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueDefault,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}
