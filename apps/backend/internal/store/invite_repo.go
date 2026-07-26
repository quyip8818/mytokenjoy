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
	ID           uuid.UUID
	CompanyID    uuid.UUID
	MemberID     uuid.UUID // the member this invite belongs to
	Email        string
	Phone        string
	UserID       uuid.UUID
	Role         string
	InviteCode   string
	ExpiresAt    time.Time
	AcceptedAt   *time.Time
	AcceptedMeta map[string]any // {"ip": "...", "ua": "..."}
	CreatedAt    time.Time
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, invite CompanyInvite) error
	GetInviteByCode(ctx context.Context, inviteCode string) (*CompanyInvite, error)
	GetInviteByMemberID(ctx context.Context, memberID uuid.UUID) (*CompanyInvite, error)
	MarkInviteAccepted(ctx context.Context, id uuid.UUID, acceptedAt time.Time, meta map[string]any) error
	UpdateInviteExpiry(ctx context.Context, id uuid.UUID, expiresAt time.Time) error
	FindPendingInvitesForUser(ctx context.Context, email string, phone string, userID uuid.UUID) ([]CompanyInvite, error)
}
