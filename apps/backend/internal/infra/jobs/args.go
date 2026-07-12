package jobs

import (
	"encoding/json"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/tokenjoy/backend/internal/config"
)

const (
	KindWalletSync               = "wallet_sync"
	KindRebalance                = "rebalance"
	KindOverrun                  = "overrun"
	KindNewAPISync               = "newapi_sync"
	KindOrgSync                  = "org_sync"
	KindMonthlyRebalance         = "monthly_rebalance"
	KindBudgetProjection         = "budget_projection"
	KindBudgetReconcile          = "budget_reconcile"
	KindBudgetReconcileFanout    = "budget_reconcile_fanout"
	KindDashboardProject         = "dashboard_project"
	KindDashboardProjectFanout   = "dashboard_project_fanout"
	KindDashboardReconcile       = "dashboard_reconcile"
	KindDashboardReconcileFanout = "dashboard_reconcile_fanout"
)

type WalletSyncArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
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

type RebalanceArgs struct {
	CompanyID int64  `json:"company_id" river:"unique"`
	AxisKind  string `json:"axis_kind" river:"unique"`
	AxisID    string `json:"axis_id" river:"unique"`
}

func (RebalanceArgs) Kind() string { return KindRebalance }

func (RebalanceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type OverrunArgs struct {
	CompanyID int64           `json:"company_id" river:"unique"`
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

type OrgSyncArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
}

// OrgSyncFanoutCompanyID marks a periodic fanout job (not a tenant).
const OrgSyncFanoutCompanyID int64 = 0

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

func OrgSyncFanoutInsertOpts() *river.InsertOpts {
	return &river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 5 * time.Minute,
		},
	}
}

type MonthlyRebalanceArgs struct{}

func (MonthlyRebalanceArgs) Kind() string { return KindMonthlyRebalance }

func (MonthlyRebalanceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: time.Minute,
		},
	}
}

type BudgetProjectionArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
}

func (BudgetProjectionArgs) Kind() string { return KindBudgetProjection }

func (BudgetProjectionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: time.Second,
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

type BudgetReconcileArgs struct {
	CompanyID int64 `json:"company_id" river:"unique"`
}

func (BudgetReconcileArgs) Kind() string { return KindBudgetReconcile }

func (BudgetReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 30 * time.Minute,
		},
	}
}

type BudgetReconcileFanoutArgs struct{}

func (BudgetReconcileFanoutArgs) Kind() string { return KindBudgetReconcileFanout }

func (BudgetReconcileFanoutArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 30 * time.Minute,
		},
	}
}

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

type DashboardProjectFanoutArgs struct{}

func (DashboardProjectFanoutArgs) Kind() string { return KindDashboardProjectFanout }

func (DashboardProjectFanoutArgs) InsertOpts() river.InsertOpts {
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

type DashboardReconcileFanoutArgs struct{}

func (DashboardReconcileFanoutArgs) Kind() string { return KindDashboardReconcileFanout }

func (DashboardReconcileFanoutArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: config.RiverQueueLow,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 24 * time.Hour,
		},
	}
}
