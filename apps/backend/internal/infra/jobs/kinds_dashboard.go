package jobs

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type DashboardProjectArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
}

func (DashboardProjectArgs) Kind() string { return KindDashboardProject }

func (DashboardProjectArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: time.Hour,
		},
	}
}

type DashboardReconcileArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
}

func (DashboardReconcileArgs) Kind() string { return KindDashboardReconcile }

func (DashboardReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 24 * time.Hour,
		},
	}
}
