package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
)

// pgListener wraps a dedicated pgxpool.Conn for LISTEN/NOTIFY.
type pgListener struct {
	pool *pgxpool.Pool
	conn *pgxpool.Conn
}

// NewPGListener acquires a connection from the pool and returns a PGListener.
func NewPGListener(ctx context.Context, pool *pgxpool.Pool) (store.PGListener, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &pgListener{pool: pool, conn: conn}, nil
}

func (l *pgListener) Listen(ctx context.Context, channel string) error {
	if err := store.ValidPGNotifyChannel(channel); err != nil {
		return err
	}
	_, err := l.conn.Exec(ctx, "LISTEN "+channel)
	return err
}

func (l *pgListener) WaitForNotification(ctx context.Context) error {
	_, err := l.conn.Conn().WaitForNotification(ctx)
	return err
}

func (l *pgListener) Close(_ context.Context) error {
	l.conn.Release()
	return nil
}
