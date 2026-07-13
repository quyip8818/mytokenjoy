package jobs

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type NewAPISyncArgs struct {
	CompanyID     int64  `json:"company_id"`
	SubKind       string `json:"sub_kind"`
	PlatformKeyID string `json:"platform_key_id,omitempty"`
	ProviderKeyID string `json:"provider_key_id,omitempty"`
	DepartmentID  string `json:"department_id,omitempty"`
}

func (NewAPISyncArgs) Kind() string { return KindNewAPISync }

func (NewAPISyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueCritical,
	}
}
