package postgres

import (
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgRepo struct {
	db    dbQuerier
	nodes *pgOrgNodeRepo
}

func (r *pgOrgRepo) Nodes() store.OrgNodeRepository {
	return r.nodes
}
