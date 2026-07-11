package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	MappingSyncStatusPending = "pending"
	MappingSyncStatusSynced  = "synced"
	MappingSyncStatusFailed  = "failed"
)

func HashPlatformKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

type PlatformKeyMapping struct {
	CompanyID            int64
	PlatformKeyID        string
	NewAPIKeyID          *int64
	MemberID             *string
	DepartmentID         string
	BudgetGroupID        *string
	NewAPIGroup          string
	SyncStatus           string
	SyncedAt             *time.Time
	NewAPIKeyRemainQuota *int64
}

type PlatformKeyMappingRepository interface {
	GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*PlatformKeyMapping, error)
	GetMappingByKeyHash(ctx context.Context, keyHash string) (*PlatformKeyMapping, error)
	FindMappingByNewAPIKeyID(ctx context.Context, keyID int64) (*PlatformKeyMapping, error)
	ListMappingsByMemberID(ctx context.Context, memberID string) ([]PlatformKeyMapping, error)
	ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]PlatformKeyMapping, error)
	ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]PlatformKeyMapping, error)
	ListActiveMappings(ctx context.Context) ([]PlatformKeyMapping, error)
	ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]PlatformKeyMapping, error)
	UpsertMapping(ctx context.Context, mapping PlatformKeyMapping) error
	UpdateMappingSync(ctx context.Context, platformKeyID string, keyID int64, status string, remainQuota *int64, syncedAt time.Time) error
	UpdateMappingNewAPIKeyRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error
}
