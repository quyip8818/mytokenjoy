package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

type RebalanceArgs struct {
	CompanyID uuid.UUID `json:"company_id" river:"unique"`
	AxisKind  string    `json:"axis_kind" river:"unique"`
	AxisID    uuid.UUID `json:"axis_id" river:"unique"`
}

func (RebalanceArgs) Kind() string { return KindRebalance }

func (RebalanceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 30 * time.Second,
		},
	}
}

type OverrunArgs struct {
	CompanyID uuid.UUID       `json:"company_id" river:"unique"`
	Payload   json.RawMessage `json:"payload" river:"unique"`
}

func (OverrunArgs) Kind() string { return KindOverrun }

func (OverrunArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type BudgetReconcileArgs struct {
	CompanyID uuid.UUID `json:"company_id" river:"unique"`
}

func (BudgetReconcileArgs) Kind() string { return KindBudgetReconcile }

func (BudgetReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 24 * time.Hour,
		},
	}
}
