package store

import (
	"context"
	"time"
)

// PrecheckContextRow is the store DTO loaded in a single round-trip for /v1 precheck.
type PrecheckContextRow struct {
	CompanyID          int64
	CompanyStatus      string
	BalancePoint       float64
	NewAPIWalletUserID *int64

	PlatformKeyID string
	KeyStatus     string
	KeyBudget     float64
	KeyConsumed   float64

	DepartmentID string
	DeptFound    bool
	DeptBudget   float64
	DeptConsumed float64
	PeriodKey    string

	MemberID       *string
	MemberFound    bool
	MemberCap      float64
	MemberConsumed float64

	BudgetGroupID *string
	GroupBudget   float64
	GroupConsumed float64

	HasAllowlist   bool
	AllowlistTypes []string
}

type GatewayPrecheckRepository interface {
	LoadPrecheckContext(ctx context.Context, keyHash string, at time.Time) (*PrecheckContextRow, error)
}
