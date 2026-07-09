//go:build testhook

package pg

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	adminOnce sync.Once
	adminPool *pgxpool.Pool
	adminErr  error
)

func EnsureAdminPool(ctx context.Context, baseURL string) (*pgxpool.Pool, error) {
	adminOnce.Do(func() {
		adminPool, adminErr = pgxpool.New(ctx, baseURL)
	})
	return adminPool, adminErr
}
