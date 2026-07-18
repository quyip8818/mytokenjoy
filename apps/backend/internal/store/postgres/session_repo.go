package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type sessionRepo struct {
	db dbQuerier
}

func newSessionRepo(db dbQuerier) *sessionRepo {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, sess store.Session) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, member_id, company_id, token_hash, user_agent, ip, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, sess.ID, sess.UserID, sess.MemberID, sess.CompanyID, sess.TokenHash,
		sess.UserAgent, sess.IP, sess.CreatedAt, sess.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *sessionRepo) GetActive(ctx context.Context, id string) (*store.Session, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, member_id, company_id, token_hash, user_agent, ip, created_at, expires_at
		FROM sessions
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, id)
	var s store.Session
	err := row.Scan(&s.ID, &s.UserID, &s.MemberID, &s.CompanyID, &s.TokenHash,
		&s.UserAgent, &s.IP, &s.CreatedAt, &s.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active session: %w", err)
	}
	return &s, nil
}

func (r *sessionRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *sessionRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}
	return nil
}

var _ store.SessionRepository = (*sessionRepo)(nil)
