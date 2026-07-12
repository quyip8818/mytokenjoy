package gateway

import (
	"github.com/tokenjoy/backend/internal/store"
)

type WalletState struct {
	CompanyID          int64
	CompanyStatus      string
	WalletRemain       float64
	NewAPIWalletUserID *int64
}

type RoutingState struct {
	PlatformKeyID  string
	KeyStatus      string
	HasAllowlist   bool
	AllowlistTypes []string
}

type PrecheckContext struct {
	Wallet  WalletState
	Routing RoutingState
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
			CompanyID:          row.CompanyID,
			CompanyStatus:      row.CompanyStatus,
			WalletRemain:       row.WalletRemain,
			NewAPIWalletUserID: row.NewAPIWalletUserID,
		},
		Routing: RoutingState{
			PlatformKeyID:  row.PlatformKeyID,
			KeyStatus:      row.KeyStatus,
			HasAllowlist:   row.HasAllowlist,
			AllowlistTypes: allowlist,
		},
	}
}
