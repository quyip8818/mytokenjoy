package store

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sms/backend/internal/domain/types"
)

const pgUniqueViolation = "23505"

func WrapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return types.ErrNotFound
	}
	return err
}

func WrapConflict(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return types.ErrConflict
	}
	return err
}
