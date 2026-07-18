package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

type platformRepo struct {
	db dbQuerier
}

func newPlatformRepo(db dbQuerier) *platformRepo {
	return &platformRepo{db: db}
}

func (r *platformRepo) GetOperatorByEmail(ctx context.Context, email string) (*store.PlatformOperator, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM platform_operators WHERE email = $1
	`, email)
	return scanPlatformOperator(row)
}

func (r *platformRepo) GetOperatorByID(ctx context.Context, id uuid.UUID) (*store.PlatformOperator, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM platform_operators WHERE id = $1
	`, id)
	return scanPlatformOperator(row)
}

func (r *platformRepo) CreateOperator(ctx context.Context, op store.PlatformOperator) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO platform_operators (id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, op.ID, op.Email, op.PasswordHash, op.Status, op.CreatedAt, op.UpdatedAt)
	return err
}

func (r *platformRepo) CountOperators(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM platform_operators`).Scan(&count)
	return count, err
}

func scanPlatformOperator(row scannable) (*store.PlatformOperator, error) {
	var op store.PlatformOperator
	if err := row.Scan(&op.ID, &op.Email, &op.PasswordHash, &op.Status, &op.CreatedAt, &op.UpdatedAt); err != nil {
		return nil, err
	}
	return &op, nil
}

var _ store.PlatformRepository = (*platformRepo)(nil)
