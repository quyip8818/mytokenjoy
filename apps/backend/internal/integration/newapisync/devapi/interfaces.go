package devapi

import (
	"context"

	"github.com/google/uuid"
)

type BearerResolver interface {
	ResolvePlatformKeyBearer(ctx context.Context, platformKeyID uuid.UUID) (string, error)
}

type ReadinessChecker interface {
	UnreadyPlatformKeyIDs(ctx context.Context) ([]uuid.UUID, error)
}
