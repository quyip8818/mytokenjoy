package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PlatformQueryRepository provides cross-company read-only aggregation queries.
// Used exclusively by the platform admin handler (SaaS mode).
type PlatformQueryRepository interface {
	SumMonthlyCost(ctx context.Context, from, to time.Time) (map[uuid.UUID]float64, error)
	CountActiveMembers(ctx context.Context) (map[uuid.UUID]int, error)
}
