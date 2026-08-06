package store

import (
	"context"

	"github.com/google/uuid"
)

// GlobalSyncVersion is the sentinel company_id for global (non-company-scoped) versions.
var GlobalSyncVersion = uuid.Nil // 00000000-0000-0000-0000-000000000000

// SyncVersionRepository manages catalog sync version counters.
// Global versions use GlobalSyncVersion as company_id; per-company versions use the real company UUID.
type SyncVersionRepository interface {
	// Increment atomically bumps version+1, returns new value. Upserts if absent.
	Increment(ctx context.Context, companyID uuid.UUID, typ string) (int, error)
	// Set writes an exact version value (upsert). Used by Local sync executor.
	Set(ctx context.Context, companyID uuid.UUID, typ string, version int) error
	// Get returns current version (0 if row does not exist).
	Get(ctx context.Context, companyID uuid.UUID, typ string) (int, error)
	// GetVersions returns global + per-company versions in one query.
	GetVersions(ctx context.Context, companyID uuid.UUID) (global map[string]int, company map[string]int, err error)
}
