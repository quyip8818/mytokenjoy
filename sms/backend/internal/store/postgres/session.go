package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	var u types.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, real_name, email, role, status, created_at, updated_at
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RealName, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	return &u, store.WrapNotFound(err)
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	var u types.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, real_name, email, role, status, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RealName, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	return &u, store.WrapNotFound(err)
}

func (s *Store) CreateSession(ctx context.Context, session *types.Session) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		session.Token, session.UserID, session.ExpiresAt,
	)
	return err
}

func (s *Store) GetSession(ctx context.Context, token string) (*types.Session, error) {
	var sess types.Session
	err := s.pool.QueryRow(ctx,
		`SELECT token, user_id, expires_at, created_at FROM sessions WHERE token = $1`, token,
	).Scan(&sess.Token, &sess.UserID, &sess.ExpiresAt, &sess.CreatedAt)
	return &sess, store.WrapNotFound(err)
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

func (s *Store) RotateSession(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET token = $1, expires_at = $2 WHERE token = $3`,
		newToken, expiresAt, oldToken,
	)
	return err
}
