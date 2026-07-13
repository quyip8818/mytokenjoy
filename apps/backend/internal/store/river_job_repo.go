package store

import "context"

// RiverJobKindOrgSync mirrors jobs.KindOrgSync; kept in store to avoid infra imports.
const RiverJobKindOrgSync = "org_sync"

type RiverJobRepository interface {
	ListActiveOrgSyncJobIDs(ctx context.Context, companyID int64) ([]int64, error)
	ListCancellableOrgSyncJobIDs(ctx context.Context, companyID int64) ([]int64, error)
	HasActiveOrgSync(ctx context.Context, companyID int64) (bool, error)
}
