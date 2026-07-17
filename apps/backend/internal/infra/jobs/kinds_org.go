package jobs

import (
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/tokenjoy/backend/internal/config"
)

type OrgSyncArgs struct {
	CompanyID uuid.UUID `json:"company_id" river:"unique"`
}

func (OrgSyncArgs) Kind() string { return KindOrgSync }

func (OrgSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}
