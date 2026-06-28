package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type credentialRepo struct {
	db dbQuerier
}

func (r *credentialRepo) GetCredential() (*types.StoredCredential, error) {
	ctx := context.Background()
	var platform string
	var encrypted []byte
	err := r.db.QueryRow(ctx, `
		SELECT platform, encrypted FROM datasource_credentials WHERE id = 1
	`).Scan(&platform, &encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &types.StoredCredential{
		Platform:  types.Platform(platform),
		Encrypted: encrypted,
	}, nil
}

func (r *credentialRepo) SaveCredential(platform types.Platform, encrypted []byte) error {
	ctx := context.Background()
	_, err := r.db.Exec(ctx, `
		INSERT INTO datasource_credentials (id, platform, encrypted, updated_at)
		VALUES (1, $1, $2, NOW())
		ON CONFLICT (id) DO UPDATE
		SET platform = EXCLUDED.platform,
		    encrypted = EXCLUDED.encrypted,
		    updated_at = NOW()
	`, string(platform), encrypted)
	return err
}

func (r *credentialRepo) ClearCredential() error {
	ctx := context.Background()
	_, err := r.db.Exec(ctx, `DELETE FROM datasource_credentials WHERE id = 1`)
	return err
}

type schedulerLockRepo struct {
	db dbQuerier
}

func (r *schedulerLockRepo) TryAcquire(
	ctx context.Context,
	lockName, holder string,
	lease time.Duration,
) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO scheduler_locks (lock_name, holder, lease_until, updated_at)
		VALUES ($1, $2, NOW() + ($3::bigint * interval '1 microsecond'), NOW())
		ON CONFLICT (lock_name) DO UPDATE
		SET holder = EXCLUDED.holder,
		    lease_until = EXCLUDED.lease_until,
		    updated_at = NOW()
		WHERE scheduler_locks.lease_until < NOW() OR scheduler_locks.holder = EXCLUDED.holder
	`, lockName, holder, lease.Microseconds())
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *schedulerLockRepo) Release(ctx context.Context, lockName, holder string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM scheduler_locks
		WHERE lock_name = $1 AND holder = $2
	`, lockName, holder)
	return err
}

var _ store.CredentialRepository = (*credentialRepo)(nil)
var _ store.SchedulerLockRepository = (*schedulerLockRepo)(nil)
