package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type KeysRepository interface {
	ProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error
	PlatformKeys(ctx context.Context) ([]types.PlatformKey, error)
	SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error
	Approvals(ctx context.Context) ([]types.KeyApproval, error)
	SetApprovals(ctx context.Context, approvals []types.KeyApproval) error
}
