package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) ListActiveMemberKeys(ctx context.Context, memberID uuid.UUID) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, platformKeySelect+`
		WHERE company_id = $1 AND member_id = $2 AND scope = 'member' AND status = 'active'
		ORDER BY id
	`, companyID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		item, err := scanPlatformKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgKeysRepo) ListActiveKeysByProjectID(ctx context.Context, projectID uuid.UUID) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, platformKeySelect+`
		WHERE company_id = $1 AND project_id = $2 AND status = 'active'
		ORDER BY id
	`, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		item, err := scanPlatformKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
