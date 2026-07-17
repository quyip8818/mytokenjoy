package store

import (
	"context"

	"github.com/google/uuid"
)

type ModelAllowlistRow struct {
	OwnerType string
	OwnerID   uuid.UUID
	ModelID   uuid.UUID
}

type ModelAllowlistRepository interface {
	List(ctx context.Context, ownerType string, ownerID uuid.UUID) ([]uuid.UUID, error)
	Replace(ctx context.Context, ownerType string, ownerID uuid.UUID, modelIDs []uuid.UUID) error
	DeleteByOwner(ctx context.Context, ownerType string, ownerID uuid.UUID) error
	IsAllowed(ctx context.Context, ownerType string, ownerID uuid.UUID, modelID uuid.UUID) (bool, error)
	IsCallTypeAllowed(ctx context.Context, ownerType string, ownerID uuid.UUID, callType string) (bool, error)
	HasAny(ctx context.Context, ownerType string, ownerID uuid.UUID) (bool, error)
}
