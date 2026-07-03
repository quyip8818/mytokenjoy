package memory

import (
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgRepo struct {
	store *Store
	nodes *memoryOrgNodeRepo
}

func (r *memoryOrgRepo) Nodes() store.OrgNodeRepository {
	return r.nodes
}
