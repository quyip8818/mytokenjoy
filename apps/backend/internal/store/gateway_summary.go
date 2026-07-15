package store

import (
	"context"
	"time"
)

type GatewaySoftSummary struct {
	PlatformKeyID string
	KeyHash       string
	SoftRemain    float64
	UpdatedAt     time.Time
	Version       int64
}

type GatewaySoftSummaryUpdate struct {
	PlatformKeyID string
	SoftRemain    float64
}

type GatewaySoftSummaryRepository interface {
	UpdateBatch(ctx context.Context, updates []GatewaySoftSummaryUpdate) ([]GatewaySoftSummary, error)
	DecrementBatch(ctx context.Context, decrements map[string]float64) ([]GatewaySoftSummary, error)
	ListByPlatformKeyIDs(ctx context.Context, keyIDs []string) ([]GatewaySoftSummary, error)
}
