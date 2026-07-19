package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CombinedKeySummary struct {
	PlatformKeyID uuid.UUID
	KeyHash       string
	Remain        int64
	UpdatedAt     time.Time
	Version       int64
}

type CombinedKeySummaryUpdate struct {
	PlatformKeyID uuid.UUID
	Remain        int64
}

type CombinedKeySummaryRepository interface {
	UpdateBatch(ctx context.Context, updates []CombinedKeySummaryUpdate) ([]CombinedKeySummary, error)
	DecrementBatch(ctx context.Context, decrements map[uuid.UUID]int64) ([]CombinedKeySummary, error)
	ListByPlatformKeyIDs(ctx context.Context, keyIDs []uuid.UUID) ([]CombinedKeySummary, error)
	// LockPlatformKeysForUpdate acquires row-level locks on the given platform_keys rows
	// to prevent concurrent absolute-value overwrites from racing with DecrementBatch.
	LockPlatformKeysForUpdate(ctx context.Context, keyIDs []uuid.UUID) error
}
