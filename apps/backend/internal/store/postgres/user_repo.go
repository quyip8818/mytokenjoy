package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type userRepo struct {
	db dbQuerier
}

func newUserRepo(db dbQuerier) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user store.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, phone, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, nilIfEmpty(user.Phone), nilIfEmpty(user.Email),
		nilIfEmpty(user.PasswordHash), user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*store.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(phone,''), COALESCE(email,''), COALESCE(password_hash,''), status, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*store.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(phone,''), COALESCE(email,''), COALESCE(password_hash,''), status, created_at, updated_at
		FROM users WHERE phone = $1
	`, phone)
	return scanUser(row)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*store.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(phone,''), COALESCE(email,''), COALESCE(password_hash,''), status, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *userRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, id, passwordHash)
	return err
}

func (r *userRepo) UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET phone = $2, updated_at = NOW() WHERE id = $1`, id, nilIfEmpty(phone))
	return err
}

func (r *userRepo) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET email = $2, updated_at = NOW() WHERE id = $1`, id, nilIfEmpty(email))
	return err
}

func (r *userRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *userRepo) HasAnyMember(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM members WHERE user_id = $1)`, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *userRepo) ListMemberCompanies(ctx context.Context, userID uuid.UUID) ([]store.MemberCompany, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.company_id, c.name,
			COALESCE((
				SELECT ro.name FROM member_roles mr
				JOIN roles ro ON ro.company_id = mr.company_id AND ro.id = mr.role_id
				WHERE mr.company_id = m.company_id AND mr.member_id = m.id
				LIMIT 1
			), '') AS role_name
		FROM members m
		JOIN companies c ON c.id = m.company_id
		WHERE m.user_id = $1 AND m.status = 'active' AND c.status = 'active'
		ORDER BY m.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []store.MemberCompany
	for rows.Next() {
		var mc store.MemberCompany
		if err := rows.Scan(&mc.MemberID, &mc.CompanyID, &mc.CompanyName, &mc.Role); err != nil {
			return nil, err
		}
		result = append(result, mc)
	}
	return result, rows.Err()
}

func scanUser(row pgx.Row) (*store.User, error) {
	var u store.User
	err := row.Scan(&u.ID, &u.Phone, &u.Email, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// nilIfEmpty returns nil for empty strings (for nullable DB columns).
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nilUUID returns nil for uuid.Nil (for nullable UUID DB columns).
func nilUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

var _ store.UserRepository = (*userRepo)(nil)
