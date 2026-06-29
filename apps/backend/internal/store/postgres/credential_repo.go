package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type credentialRepo struct {
	ctx context.Context
	db  dbQuerier
}

func (r *credentialRepo) GetCredential() (*types.StoredCredential, error) {
	var platform string
	var encrypted []byte
	err := r.db.QueryRow(r.ctx, `
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
	_, err := r.db.Exec(r.ctx, `
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
	_, err := r.db.Exec(r.ctx, `DELETE FROM datasource_credentials WHERE id = 1`)
	return err
}

var _ store.CredentialRepository = (*credentialRepo)(nil)
