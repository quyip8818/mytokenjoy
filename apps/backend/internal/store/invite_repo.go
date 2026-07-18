package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	InviteRoleSuperAdmin = "super_admin"
)

type CompanyInvite struct {
	ID         uuid.UUID
	CompanyID  uuid.UUID
	Email      string
	Phone      string
	UserID     uuid.UUID
	Role       string
	InviteCode string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, invite CompanyInvite) error
	GetInviteByCode(ctx context.Context, inviteCode string) (*CompanyInvite, error)
	MarkInviteAccepted(ctx context.Context, id uuid.UUID, acceptedAt time.Time) error
	FindPendingInvitesForUser(ctx context.Context, email string, phone string, userID uuid.UUID) ([]CompanyInvite, error)
}
