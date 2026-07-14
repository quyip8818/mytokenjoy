package store

import (
	"context"
	"time"
)

// PrecheckContextRow is the store DTO loaded in a single round-trip for /v1 precheck.
type PrecheckContextRow struct {
	CompanyID          int64
	CompanyStatus      string
	WalletRemain       float64
	NewAPIWalletUserID *int64

	PlatformKeyID string
	KeyStatus     string
	KeyExpiresAt  *time.Time

	HasAllowlist   bool
	AllowlistTypes []string

	GatewaySoftRemain  *float64
	GatewaySoftAt      *time.Time
	GatewaySoftVersion int64
}

type GatewayPrecheckRepository interface {
	LoadPrecheckContext(ctx context.Context, keyHash string) (*PrecheckContextRow, error)
}
