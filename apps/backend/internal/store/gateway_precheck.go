package store

import (
	"context"
	"time"
)

// PrecheckContextRow is the store DTO loaded in a single round-trip for /v1 precheck.
type PrecheckContextRow struct {
	CompanyID     int64
	CompanyType   string
	CompanyStatus string
	WalletRemain  float64

	PlatformKeyID string
	KeyStatus     string
	KeyExpiresAt  *time.Time

	HasAllowlist   bool
	AllowlistTypes []string

	CombinedKeyRemain        *float64
	CombinedKeyRemainAt      *time.Time
	CombinedKeyRemainVersion int64
}

type GatewayPrecheckRepository interface {
	LoadPrecheckContext(ctx context.Context, keyHash string) (*PrecheckContextRow, error)
}
