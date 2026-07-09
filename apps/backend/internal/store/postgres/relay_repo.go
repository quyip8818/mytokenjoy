package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/store"
)

type dbQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type relayRepo struct {
	db dbQuerier
}

func newRelayRepo(db dbQuerier) *relayRepo {
	return &relayRepo{db: db}
}

var _ store.RelayRepository = (*relayRepo)(nil)
