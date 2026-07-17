package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type KeysRepository interface {
	ProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error
	PlatformKeys(ctx context.Context) ([]types.PlatformKey, error)
	PlatformKeyByID(ctx context.Context, keyID uuid.UUID) (*types.PlatformKey, error)
	PlatformKeyByHash(ctx context.Context, keyHash string) (*types.PlatformKey, error)
	PlatformKeyHashByID(ctx context.Context, keyID uuid.UUID) (string, bool, error)
	SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error
	ListActiveMemberKeys(ctx context.Context, memberID uuid.UUID) ([]types.PlatformKey, error)
	ListActiveKeysByProjectID(ctx context.Context, projectID uuid.UUID) ([]types.PlatformKey, error)
	Approvals(ctx context.Context) ([]types.KeyApproval, error)
	SetApprovals(ctx context.Context, approvals []types.KeyApproval) error
}
