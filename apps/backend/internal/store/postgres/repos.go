package postgres

import (
	"github.com/tokenjoy/backend/internal/store"
)

func newDomainRepoSet(db dbQuerier) domainRepos {
	return domainRepos{
		org:    &pgOrgRepo{db: db},
		budget: &pgBudgetRepo{db: db},
		keys:   &pgKeysRepo{db: db},
		models: &pgModelsRepo{db: db},
		audit:  &pgAuditRepo{db: db},
	}
}

var (
	_ store.OrgRepository    = (*pgOrgRepo)(nil)
	_ store.BudgetRepository = (*pgBudgetRepo)(nil)
	_ store.KeysRepository   = (*pgKeysRepo)(nil)
	_ store.ModelsRepository = (*pgModelsRepo)(nil)
	_ store.AuditRepository  = (*pgAuditRepo)(nil)
)
