//go:build testhook

package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
)

const SaaSMinCompanyIDForTest int64 = saasMinCompanyID

func EnsureBootstrapCompanyForTest(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	return ensureBootstrapCompany(ctx, pool, cfg)
}
