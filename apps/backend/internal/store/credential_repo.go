package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type CredentialRepository interface {
	GetCredential(ctx context.Context) (*types.StoredCredential, error)
	SaveCredential(ctx context.Context, platform types.Platform, encrypted []byte) error
	ClearCredential(ctx context.Context) error
}

type SchedulerLockRepository interface {
	TryAcquire(ctx context.Context, lockName, holder string, lease time.Duration) (bool, error)
	Release(ctx context.Context, lockName, holder string) error
}
