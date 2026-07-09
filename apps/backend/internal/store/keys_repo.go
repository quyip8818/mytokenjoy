package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type KeysRepository interface {
	ProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error
	PlatformKeys(ctx context.Context) ([]types.PlatformKey, error)
	PlatformKeyByID(ctx context.Context, keyID string) (*types.PlatformKey, error)
	PlatformKeyByHash(ctx context.Context, keyHash string) (*types.PlatformKey, error)
	SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error
	SumMemberKeyUsed(ctx context.Context, memberID string, at time.Time) (float64, error)
	ListActiveMemberKeys(ctx context.Context, memberID string) ([]types.PlatformKey, error)
	Approvals(ctx context.Context) ([]types.KeyApproval, error)
	SetApprovals(ctx context.Context, approvals []types.KeyApproval) error
}
