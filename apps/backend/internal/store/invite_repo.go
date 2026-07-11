package store

import (
	"context"
	"time"
)

const (
	InviteRoleSuperAdmin = "super_admin"
)

type CompanyInvite struct {
	ID         string
	CompanyID  int64
	Email      string
	Role       string
	Token      string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, invite CompanyInvite) error
	GetInviteByToken(ctx context.Context, token string) (*CompanyInvite, error)
	MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error
}
