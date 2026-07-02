package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type KeysRepository interface {
	ProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error
	PlatformKeys(ctx context.Context) ([]types.PlatformKey, error)
	PlatformKeyByID(ctx context.Context, keyID string) (*types.PlatformKey, error)
	SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error
	AddPlatformKeyUsed(ctx context.Context, keyID string, amountCNY float64) error
	SumMemberKeyUsed(ctx context.Context, memberID string) (float64, error)
	ListActiveMemberKeys(ctx context.Context, memberID string) ([]types.PlatformKey, error)
	Approvals(ctx context.Context) ([]types.KeyApproval, error)
	SetApprovals(ctx context.Context, approvals []types.KeyApproval) error
}
