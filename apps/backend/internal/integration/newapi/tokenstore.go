package newapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

// TokenStore reads the admin access_token directly from NewAPI's Postgres database.
type TokenStore struct {
	dsn         string
	adminUserID int64
}

// NewTokenStore creates a TokenStore that reads from the given NewAPI database DSN.
func NewTokenStore(dsn string, adminUserID int64) *TokenStore {
	return &TokenStore{dsn: dsn, adminUserID: adminUserID}
}

// FetchToken reads the access_token for the configured admin user from NewAPI's users table.
func (ts *TokenStore) FetchToken(ctx context.Context) (string, error) {
	conn, err := pgx.Connect(ctx, ts.dsn)
	if err != nil {
		return "", fmt.Errorf("connect newapi db: %w", err)
	}
	defer conn.Close(ctx)

	var token string
	err = conn.QueryRow(ctx,
		`SELECT access_token FROM users WHERE id = $1`,
		ts.adminUserID).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("read admin token from newapi db: %w", err)
	}
	if token == "" {
		return "", fmt.Errorf("admin user %d has empty access_token", ts.adminUserID)
	}
	return token, nil
}

// DeriveNewAPIDSN derives the NewAPI database DSN from the main DATABASE_URL by
// replacing the database name with "newapi".
func DeriveNewAPIDSN(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	u.Path = "/newapi"
	return u.String(), nil
}
