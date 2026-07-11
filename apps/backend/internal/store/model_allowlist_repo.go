package store

import "context"

type ModelAllowlistRow struct {
	OwnerType string
	OwnerID   string
	ModelID   int64
}

type ModelAllowlistRepository interface {
	List(ctx context.Context, ownerType, ownerID string) ([]int64, error)
	Replace(ctx context.Context, ownerType, ownerID string, modelIDs []int64) error
	DeleteByOwner(ctx context.Context, ownerType, ownerID string) error
	IsAllowed(ctx context.Context, ownerType, ownerID string, modelID int64) (bool, error)
	IsCallTypeAllowed(ctx context.Context, ownerType, ownerID, callType string) (bool, error)
	HasAny(ctx context.Context, ownerType, ownerID string) (bool, error)
}
