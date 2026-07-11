package gateway

import (
	"github.com/tokenjoy/backend/internal/store"
)

type WalletState struct {
	CompanyID          int64
	CompanyStatus      string
	BalancePoint       float64
	NewAPIWalletUserID *int64
}

type BudgetState struct {
	DepartmentID string
	DeptFound    bool
	DeptBudget   float64
	DeptConsumed float64
	PeriodKey    string

	PlatformKeyID string
	KeyBudget     float64
	KeyConsumed   float64

	MemberID       *string
	MemberFound    bool
	MemberCap      float64
	MemberConsumed float64

	BudgetGroupID *string
	GroupBudget   float64
	GroupConsumed float64
}

// PolicyState holds org-level policy knobs; Phase 4 adds org_budget_mode.
type PolicyState struct{}

type RoutingState struct {
	PlatformKeyID  string
	KeyStatus      string
	HasAllowlist   bool
	AllowlistTypes []string
}

type PrecheckContext struct {
	Wallet  WalletState
	Budget  BudgetState
	Policy  PolicyState
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
			BalancePoint:       row.BalancePoint,
			NewAPIWalletUserID: row.NewAPIWalletUserID,
		},
		Budget: BudgetState{
			DepartmentID:   row.DepartmentID,
			DeptFound:      row.DeptFound,
			DeptBudget:     row.DeptBudget,
			DeptConsumed:   row.DeptConsumed,
			PeriodKey:      row.PeriodKey,
			PlatformKeyID:  row.PlatformKeyID,
			KeyBudget:      row.KeyBudget,
			KeyConsumed:    row.KeyConsumed,
			MemberID:       row.MemberID,
			MemberFound:    row.MemberFound,
			MemberCap:      row.MemberCap,
			MemberConsumed: row.MemberConsumed,
			BudgetGroupID:  row.BudgetGroupID,
			GroupBudget:    row.GroupBudget,
			GroupConsumed:  row.GroupConsumed,
		},
		Policy: PolicyState{},
		Routing: RoutingState{
			PlatformKeyID:  row.PlatformKeyID,
			KeyStatus:      row.KeyStatus,
			HasAllowlist:   row.HasAllowlist,
			AllowlistTypes: allowlist,
		},
	}
}
