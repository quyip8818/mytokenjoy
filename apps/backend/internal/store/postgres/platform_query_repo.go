package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
)

type platformQueryRepo struct {
	db *pgxpool.Pool
}

func newPlatformQueryRepo(pool *pgxpool.Pool) *platformQueryRepo {
	return &platformQueryRepo{db: pool}
}

func (r *platformQueryRepo) SumMonthlyCost(ctx context.Context, from, to time.Time) (map[uuid.UUID]float64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT company_id, COALESCE(SUM(cost), 0)
		FROM usage_buckets
		WHERE bucket_start >= $1 AND bucket_start < $2
		GROUP BY company_id
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[uuid.UUID]float64)
	for rows.Next() {
		var id uuid.UUID
		var cost float64
		if err := rows.Scan(&id, &cost); err != nil {
			return nil, err
		}
		result[id] = cost
	}
	return result, rows.Err()
}

func (r *platformQueryRepo) CountActiveMembers(ctx context.Context) (map[uuid.UUID]int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT company_id, COUNT(*)
		FROM members
		WHERE status = 'active'
		GROUP BY company_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[uuid.UUID]int)
	for rows.Next() {
		var id uuid.UUID
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		result[id] = count
	}
	return result, rows.Err()
}

var _ store.PlatformQueryRepository = (*platformQueryRepo)(nil)
