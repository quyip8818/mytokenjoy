package postgres

import (
	"context"

	"github.com/google/uuid"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/store"
)

func (s *Store) ListUsers(ctx context.Context) ([]types.User, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, username, real_name, email, role, status, created_at, updated_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.User
	for rows.Next() {
		var u types.User
		if err := rows.Scan(&u.ID, &u.Username, &u.RealName, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	if items == nil {
		items = []types.User{}
	}
	return items, nil
}

func (s *Store) CreateUser(ctx context.Context, u *types.User) error {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (id, username, password_hash, real_name, email, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`,
		u.ID, u.Username, u.PasswordHash, u.RealName, u.Email, u.Role, u.Status,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	return store.WrapConflict(err)
}

func (s *Store) UpdateUser(ctx context.Context, u *types.User) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE users SET real_name=$1, email=$2, role=$3, password_hash=$4, status=$5 WHERE id=$6`,
		u.RealName, u.Email, u.Role, u.PasswordHash, u.Status, u.ID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}
