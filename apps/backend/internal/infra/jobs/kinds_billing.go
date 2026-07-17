package jobs

import (
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type WalletSyncArgs struct {
	CompanyID uuid.UUID `json:"company_id" river:"unique"`
}

func (WalletSyncArgs) Kind() string { return KindWalletSync }

func (WalletSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 5 * time.Second,
		},
	}
}
