package postgres

import (
	"context"
	"fmt"

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

func (r *userRepo) GetByID(ctx context.Context, id string) (*store.User, error) {
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

func (r *userRepo) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, id, passwordHash)
	return err
}

func (r *userRepo) UpdatePhone(ctx context.Context, id string, phone string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET phone = $2, updated_at = NOW() WHERE id = $1`, id, nilIfEmpty(phone))
	return err
}

func (r *userRepo) UpdateEmail(ctx context.Context, id string, email string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET email = $2, updated_at = NOW() WHERE id = $1`, id, nilIfEmpty(email))
	return err
}

func (r *userRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
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

var _ store.UserRepository = (*userRepo)(nil)
