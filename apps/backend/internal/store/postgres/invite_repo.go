package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

// nilIfNilUUID returns nil for uuid.Nil (for nullable DB columns).
func nilIfNilUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

type inviteRepo struct {
	db dbQuerier
}

func newInviteRepo(db dbQuerier) *inviteRepo {
	return &inviteRepo{db: db}
}

func (r *inviteRepo) CreateInvite(ctx context.Context, invite store.CompanyInvite) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_invites (id, company_id, email, phone, user_id, role, invite_code, expires_at, accepted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, invite.ID, invite.CompanyID, nilIfEmpty(invite.Email), nilIfEmpty(invite.Phone),
		nilIfNilUUID(invite.UserID), invite.Role, invite.InviteCode,
		invite.ExpiresAt, invite.AcceptedAt, invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (r *inviteRepo) GetInviteByCode(ctx context.Context, inviteCode string) (*store.CompanyInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, COALESCE(email,''), COALESCE(phone,''), COALESCE(user_id, '00000000-0000-0000-0000-000000000000'), role, invite_code, expires_at, accepted_at, created_at
		FROM company_invites WHERE invite_code = $1
	`, inviteCode)
	var inv store.CompanyInvite
	if err := row.Scan(&inv.ID, &inv.CompanyID, &inv.Email, &inv.Phone, &inv.UserID,
		&inv.Role, &inv.InviteCode, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) MarkInviteAccepted(ctx context.Context, id uuid.UUID, acceptedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_invites SET accepted_at = $2 WHERE id = $1
	`, id, acceptedAt)
	return err
}

func (r *inviteRepo) FindPendingInvitesForUser(ctx context.Context, email string, phone string, userID uuid.UUID) ([]store.CompanyInvite, error) {
	// Dynamically build WHERE conditions based on non-empty identifiers.
	conditions := []string{}
	args := []any{}
	argIdx := 1

	if email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, email)
		argIdx++
	}
	if phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, phone)
		argIdx++
	}
	if userID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	query := `SELECT id, company_id, COALESCE(email,''), COALESCE(phone,''), COALESCE(user_id, '00000000-0000-0000-0000-000000000000'), role, invite_code, expires_at, accepted_at, created_at
		FROM company_invites
		WHERE accepted_at IS NULL AND expires_at > NOW() AND (`
	for i, cond := range conditions {
		if i > 0 {
			query += " OR "
		}
		query += cond
	}
	query += ")"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find pending invites: %w", err)
	}
	defer rows.Close()

	var invites []store.CompanyInvite
	for rows.Next() {
		var inv store.CompanyInvite
		if err := rows.Scan(&inv.ID, &inv.CompanyID, &inv.Email, &inv.Phone, &inv.UserID,
			&inv.Role, &inv.InviteCode, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

var _ store.InviteRepository = (*inviteRepo)(nil)
