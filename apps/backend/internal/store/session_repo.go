package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Session represents a refresh-token-backed login session.
type Session struct {
	ID        string
	UserID    uuid.UUID
	MemberID  uuid.UUID
	CompanyID uuid.UUID
	TokenHash string
	UserAgent string
	IP        string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// SessionRepository manages refresh token sessions.
type SessionRepository interface {
	Create(ctx context.Context, sess Session) error
	GetActive(ctx context.Context, id string) (*Session, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
}
