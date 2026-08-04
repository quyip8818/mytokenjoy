package store

import (
	"context"
	"encoding/json"
)

// BudgetSnapshot is the full-company budget configuration at a given period.
// The JSON structure mirrors the GET /budget/tree response + members + projects.
type BudgetSnapshot struct {
	PeriodKey string          `json:"periodKey"`
	Snapshot  json.RawMessage `json:"snapshot"`
}

type BudgetSnapshotRepository interface {
	// Get returns the snapshot for a period. Returns nil if not found.
	Get(ctx context.Context, periodKey string) (*BudgetSnapshot, error)
	// GetForUpdate returns the snapshot with a row-level lock (SELECT FOR UPDATE).
	GetForUpdate(ctx context.Context, periodKey string) (*BudgetSnapshot, error)
	// Upsert inserts or replaces the snapshot for a period.
	Upsert(ctx context.Context, periodKey string, snapshot json.RawMessage) error
	// Delete removes a snapshot (e.g. after rotation applied it).
	Delete(ctx context.Context, periodKey string) error
	// ListPeriods returns all period_keys that have snapshots, ordered ascending.
	ListPeriods(ctx context.Context) ([]string, error)
	// Exists checks whether a snapshot exists for the given period.
	Exists(ctx context.Context, periodKey string) (bool, error)
}
