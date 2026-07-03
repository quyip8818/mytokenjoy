package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type ModelsRepository interface {
	Models(ctx context.Context) ([]types.ModelInfo, error)
	ModelByName(ctx context.Context, name string) (*types.ModelInfo, error)
	SetModels(ctx context.Context, models []types.ModelInfo) error
	Allowlist() ModelAllowlistRepository
}
