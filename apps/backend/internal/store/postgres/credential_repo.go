package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type credentialRepo struct {
	db dbQuerier
}

func (r *credentialRepo) GetCredential(ctx context.Context) (*types.StoredCredential, error) {
	companyID := store.CompanyID(ctx)
	var platform string
	var encrypted []byte
	err := r.db.QueryRow(ctx, `
		SELECT platform, encrypted FROM datasource_credentials WHERE company_id = $1
	`, companyID).Scan(&platform, &encrypted)
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

func (r *credentialRepo) SaveCredential(ctx context.Context, platform types.Platform, encrypted []byte) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO datasource_credentials (company_id, platform, encrypted, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (company_id) DO UPDATE
		SET platform = EXCLUDED.platform,
		    encrypted = EXCLUDED.encrypted,
		    updated_at = NOW()
	`, companyID, string(platform), encrypted)
	return err
}

func (r *credentialRepo) ClearCredential(ctx context.Context) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `DELETE FROM datasource_credentials WHERE company_id = $1`, companyID)
	return err
}

var _ store.CredentialRepository = (*credentialRepo)(nil)
