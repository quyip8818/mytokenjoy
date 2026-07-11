package store

import (
	"context"
)

// PrecheckContextRow is the store DTO loaded in a single round-trip for /v1 precheck.
type PrecheckContextRow struct {
	CompanyID          int64
	CompanyStatus      string
	BalancePoint       float64
	NewAPIWalletUserID *int64

	PlatformKeyID string
	KeyStatus     string

	HasAllowlist   bool
	AllowlistTypes []string
}

type GatewayPrecheckRepository interface {
	LoadPrecheckContext(ctx context.Context, keyHash string) (*PrecheckContextRow, error)
}
