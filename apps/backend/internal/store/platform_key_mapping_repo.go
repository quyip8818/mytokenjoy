package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
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
	CompanyID     uuid.UUID
	PlatformKeyID uuid.UUID
	NewAPIKeyID   *int64
	MemberID      *uuid.UUID
	DepartmentID  uuid.UUID
	ProjectID     *uuid.UUID
	NewAPIGroup   string
	SyncStatus    string
	SyncedAt      *time.Time
}

type PlatformKeyMappingRepository interface {
	GetMappingByPlatformKeyID(ctx context.Context, platformKeyID uuid.UUID) (*PlatformKeyMapping, error)
	GetMappingByKeyHash(ctx context.Context, keyHash string) (*PlatformKeyMapping, error)
	FindMappingByNewAPIKeyID(ctx context.Context, keyID int64) (*PlatformKeyMapping, error)
	ListMappingsByNewAPIKeyIDs(ctx context.Context, keyIDs []int64) ([]PlatformKeyMapping, error)
	ListMappingsByMemberID(ctx context.Context, memberID uuid.UUID) ([]PlatformKeyMapping, error)
	ListMappingsByDepartmentID(ctx context.Context, departmentID uuid.UUID) ([]PlatformKeyMapping, error)
	ListMappingsByProjectID(ctx context.Context, projectID uuid.UUID) ([]PlatformKeyMapping, error)
	ListMappingsByPlatformKeyIDs(ctx context.Context, platformKeyIDs []uuid.UUID) ([]PlatformKeyMapping, error)
	ListActiveMappingsByCompany(ctx context.Context, companyID uuid.UUID) ([]PlatformKeyMapping, error)
	UpsertMapping(ctx context.Context, mapping PlatformKeyMapping) error
	UpdateMappingSync(ctx context.Context, platformKeyID uuid.UUID, keyID int64, status string, syncedAt time.Time) error
}
