package store

import (
	"context"

	"github.com/google/uuid"
	// RiverJobKindOrgSync mirrors jobs.KindOrgSync; kept in store to avoid infra imports.
)

const RiverJobKindOrgSync = "org_sync"

type RiverJobRepository interface {
	ListActiveOrgSyncJobIDs(ctx context.Context, companyID uuid.UUID) ([]int64, error)
	ListCancellableOrgSyncJobIDs(ctx context.Context, companyID uuid.UUID) ([]int64, error)
	HasActiveOrgSync(ctx context.Context, companyID uuid.UUID) (bool, error)
	CountActiveJobs(ctx context.Context) (int, error)
	// CountRunnableJobs is jobs workers can claim now (excludes future scheduled).
	CountRunnableJobs(ctx context.Context) (int, error)
}
