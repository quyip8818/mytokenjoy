package devapi

import "context"

type BearerResolver interface {
	ResolvePlatformKeyBearer(ctx context.Context, platformKeyID string) (string, error)
}

type ReadinessChecker interface {
	UnreadyPlatformKeyIDs(ctx context.Context) ([]string, error)
}
