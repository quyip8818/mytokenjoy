package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type riverJobRepo struct {
	db dbQuerier
}

func newRiverJobRepo(db dbQuerier) store.RiverJobRepository {
	return &riverJobRepo{db: db}
}

func (r *riverJobRepo) ListActiveOrgSyncJobIDs(ctx context.Context, companyID int64) ([]int64, error) {
	return r.listOrgSyncJobIDs(ctx, companyID, true)
}

func (r *riverJobRepo) ListCancellableOrgSyncJobIDs(ctx context.Context, companyID int64) ([]int64, error) {
	return r.listOrgSyncJobIDs(ctx, companyID, false)
}

func (r *riverJobRepo) listOrgSyncJobIDs(ctx context.Context, companyID int64, includeRunning bool) ([]int64, error) {
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

func (r *riverJobRepo) HasActiveOrgSync(ctx context.Context, companyID int64) (bool, error) {
	ids, err := r.ListActiveOrgSyncJobIDs(ctx, companyID)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

var _ store.RiverJobRepository = (*riverJobRepo)(nil)
