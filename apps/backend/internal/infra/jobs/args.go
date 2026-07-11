package jobs

import (
	"encoding/json"
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

const (
	KindWalletSync       = "wallet_sync"
	KindRebalance        = "rebalance"
	KindOverrun          = "overrun"
	KindNewAPISync       = "newapi_sync"
	KindOrgSync          = "org_sync"
	KindMonthlyRebalance = "monthly_rebalance"
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

type OrgSyncArgs struct{}

func (OrgSyncArgs) Kind() string { return KindOrgSync }

func (OrgSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: config.RiverQueueDefault}
}

type MonthlyRebalanceArgs struct{}

func (MonthlyRebalanceArgs) Kind() string { return KindMonthlyRebalance }

func (MonthlyRebalanceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: config.RiverQueueDefault}
}
