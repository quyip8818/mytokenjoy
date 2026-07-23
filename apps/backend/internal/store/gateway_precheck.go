package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PrecheckContextRow is the store DTO loaded in a single round-trip for /v1 precheck.
type PrecheckContextRow struct {
	CompanyID         uuid.UUID
	CompanyType       string
	CompanyStatus     string
	WalletQuotaRemain int64

	PlatformKeyID uuid.UUID
	KeyStatus     string
	KeyExpiresAt  *time.Time

	HasAllowlist   bool
	AllowlistTypes []string

	CombinedKeyRemain        *int64
	CombinedKeyRemainAt      *time.Time
	CombinedKeyRemainVersion int64
}

type GatewayPrecheckRepository interface {
	LoadPrecheckContext(ctx context.Context, keyHash string) (*PrecheckContextRow, error)
}
