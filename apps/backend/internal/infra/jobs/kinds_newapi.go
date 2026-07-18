package jobs

import (
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type NewAPISyncArgs struct {
	CompanyID     uuid.UUID `json:"company_id"`
	SubKind       string    `json:"sub_kind"`
	PlatformKeyID uuid.UUID `json:"platform_key_id,omitempty"`
	ProviderKeyID uuid.UUID `json:"provider_key_id,omitempty"`
	DepartmentID  uuid.UUID `json:"department_id,omitempty"`
}

func (NewAPISyncArgs) Kind() string { return KindNewAPISync }

func (NewAPISyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueCritical,
	}
}
