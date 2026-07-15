package store

import (
	"context"
	"time"
)

type CombinedKeySummary struct {
	PlatformKeyID string
	KeyHash       string
	Remain        float64
	UpdatedAt     time.Time
	Version       int64
}

type CombinedKeySummaryUpdate struct {
	PlatformKeyID string
	Remain        float64
}

type CombinedKeySummaryRepository interface {
	UpdateBatch(ctx context.Context, updates []CombinedKeySummaryUpdate) ([]CombinedKeySummary, error)
	DecrementBatch(ctx context.Context, decrements map[string]float64) ([]CombinedKeySummary, error)
	ListByPlatformKeyIDs(ctx context.Context, keyIDs []string) ([]CombinedKeySummary, error)
}
