package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type ModelsRepository interface {
	Models(ctx context.Context) ([]types.ModelInfo, error)
	ModelByType(ctx context.Context, modelType string) (*types.ModelInfo, error)
	ModelByProviderType(ctx context.Context, provider, modelType string) (*types.ModelInfo, error)
	ModelByID(ctx context.Context, modelID uuid.UUID) (*types.ModelInfo, error)
	ModelByIDs(ctx context.Context, modelIDs []int64) ([]types.ModelInfo, error)
	InsertModel(ctx context.Context, model types.ModelInfo) (types.ModelInfo, error)
	UpdateModel(ctx context.Context, model types.ModelInfo) error
	DeleteModel(ctx context.Context, modelID uuid.UUID) error
	UpsertFromSMS(ctx context.Context, modelType, name, provider, callType string) error
	DisableStaleFromSMS(ctx context.Context, activeModelIDs []string) (int, error)
	Allowlist() ModelAllowlistRepository
}
