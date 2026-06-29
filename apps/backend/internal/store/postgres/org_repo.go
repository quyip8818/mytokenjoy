package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgRepo struct {
	ctx context.Context
	db  dbQuerier
}

func (r *pgOrgRepo) Permissions() []types.Permission {
	rows, err := r.db.Query(r.ctx, `SELECT id, name, grp FROM permissions ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.Permission, 0)
	for rows.Next() {
		var item types.Permission
		if err := rows.Scan(&item.ID, &item.Name, &item.Group); err != nil {
			return nil
		}
		items = append(items, item)
	}
	return store.ClonePermissions(items)
}
