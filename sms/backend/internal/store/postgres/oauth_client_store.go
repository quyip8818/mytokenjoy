package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sms/backend/internal/domain/oauth"
)

type pgOAuthClientStore struct {
	db *pgxpool.Pool
}

func NewOAuthClientStore(db *pgxpool.Pool) *pgOAuthClientStore {
	return &pgOAuthClientStore{db: db}
}

func (s *pgOAuthClientStore) GetClientByID(ctx context.Context, clientID string) (*oauth.Client, error) {
	var client oauth.Client
	err := s.db.QueryRow(ctx, `
		SELECT client_id, client_secret_hash, scope
		FROM oauth_clients
		WHERE client_id = $1
	`, clientID).Scan(&client.ClientID, &client.ClientSecretHash, &client.Scope)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}
