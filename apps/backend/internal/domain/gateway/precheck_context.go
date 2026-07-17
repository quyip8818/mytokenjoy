package gateway

import (
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type WalletState struct {
	CompanyID     int64
	CompanyStatus string
	WalletRemain  float64
}

type RoutingState struct {
	PlatformKeyID  string
	KeyStatus      string
	KeyExpiresAt   *time.Time
	HasAllowlist   bool
	AllowlistTypes []string
}

type BudgetState struct {
	Remain  *float64
	Version int64
}

type PrecheckContext struct {
	Wallet  WalletState
	Routing RoutingState
	Budget  BudgetState
}

func PrecheckContextFromStore(row *store.PrecheckContextRow) PrecheckContext {
	if row == nil {
		return PrecheckContext{}
	}
	allowlist := row.AllowlistTypes
	if allowlist == nil {
		allowlist = []string{}
	}
	return PrecheckContext{
		Wallet: WalletState{
			CompanyID:     row.CompanyID,
			CompanyStatus: row.CompanyStatus,
			WalletRemain:  row.WalletRemain,
		},
		Routing: RoutingState{
			PlatformKeyID:  row.PlatformKeyID,
			KeyStatus:      row.KeyStatus,
			KeyExpiresAt:   row.KeyExpiresAt,
			HasAllowlist:   row.HasAllowlist,
			AllowlistTypes: allowlist,
		},
		Budget: BudgetState{
			Remain:  row.CombinedKeyRemain,
			Version: row.CombinedKeyRemainVersion,
		},
	}
}
