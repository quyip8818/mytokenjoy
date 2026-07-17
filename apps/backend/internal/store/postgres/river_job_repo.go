package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

type riverJobRepo struct {
	db dbQuerier
}

func newRiverJobRepo(db dbQuerier) store.RiverJobRepository {
	return &riverJobRepo{db: db}
}

func (r *riverJobRepo) ListActiveOrgSyncJobIDs(ctx context.Context, companyID uuid.UUID) ([]int64, error) {
	return r.listOrgSyncJobIDs(ctx, companyID, true)
}

func (r *riverJobRepo) ListCancellableOrgSyncJobIDs(ctx context.Context, companyID uuid.UUID) ([]int64, error) {
	return r.listOrgSyncJobIDs(ctx, companyID, false)
}

func (r *riverJobRepo) listOrgSyncJobIDs(ctx context.Context, companyID uuid.UUID, includeRunning bool) ([]int64, error) {
	states := `('available', 'pending', 'scheduled', 'retryable')`
	if includeRunning {
		states = `('available', 'pending', 'scheduled', 'running', 'retryable')`
	}
	query := `
		SELECT id
		FROM river_job
		WHERE kind = $2
		  AND (args->>'company_id')::bigint = $1
		  AND state IN ` + states
	rows, err := r.db.Query(ctx, query, companyID, store.RiverJobKindOrgSync)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *riverJobRepo) HasActiveOrgSync(ctx context.Context, companyID uuid.UUID) (bool, error) {
	ids, err := r.ListActiveOrgSyncJobIDs(ctx, companyID)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

func (r *riverJobRepo) CountActiveJobs(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM river_job
		WHERE state IN ('available', 'pending', 'scheduled', 'running', 'retryable')
	`).Scan(&n)
	return n, err
}

func (r *riverJobRepo) CountRunnableJobs(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM river_job
		WHERE state IN ('available', 'pending', 'running', 'retryable')
		   OR (state = 'scheduled' AND scheduled_at <= NOW())
	`).Scan(&n)
	return n, err
}

var _ store.RiverJobRepository = (*riverJobRepo)(nil)
