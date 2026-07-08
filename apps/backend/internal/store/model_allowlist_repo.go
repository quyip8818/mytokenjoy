package store

import "context"

type ModelAllowlistRepository interface {
	List(ctx context.Context, ownerType, ownerID string) ([]string, error)
	Replace(ctx context.Context, ownerType, ownerID string, models []string) error
	DeleteByOwner(ctx context.Context, ownerType, ownerID string) error
	IsAllowed(ctx context.Context, ownerType, ownerID, modelName string) (bool, error)
	HasAny(ctx context.Context, ownerType, ownerID string) (bool, error)
}
