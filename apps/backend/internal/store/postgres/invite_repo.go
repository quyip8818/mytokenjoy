package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type inviteRepo struct {
	db dbQuerier
}

func newInviteRepo(db dbQuerier) *inviteRepo {
	return &inviteRepo{db: db}
}

func (r *inviteRepo) CreateInvite(ctx context.Context, invite store.CompanyInvite) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_invites (id, company_id, email, role, invite_code, expires_at, accepted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invite.ID, invite.CompanyID, invite.Email, invite.Role, invite.InviteCode,
		invite.ExpiresAt, invite.AcceptedAt, invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (r *inviteRepo) GetInviteByCode(ctx context.Context, inviteCode string) (*store.CompanyInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, email, role, invite_code, expires_at, accepted_at, created_at
		FROM company_invites WHERE invite_code = $1
	`, inviteCode)
	var inv store.CompanyInvite
	if err := row.Scan(&inv.ID, &inv.CompanyID, &inv.Email, &inv.Role, &inv.InviteCode,
		&inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_invites SET accepted_at = $2 WHERE id = $1
	`, id, acceptedAt)
	return err
}

var _ store.InviteRepository = (*inviteRepo)(nil)
