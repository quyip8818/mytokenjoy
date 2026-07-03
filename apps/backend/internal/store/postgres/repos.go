package postgres

import (
	"github.com/tokenjoy/backend/internal/store"
)

func newDomainRepoSet(db dbQuerier) domainRepos {
	orgNodes := &pgOrgNodeRepo{db: db}
	allowlist := &pgModelAllowlistRepo{db: db}
	return domainRepos{
		org:    &pgOrgRepo{db: db, nodes: orgNodes},
		budget: &pgBudgetRepo{db: db},
		keys:   &pgKeysRepo{db: db, allowlist: allowlist},
		models: &pgModelsRepo{db: db, allowlist: allowlist},
		audit:  &pgAuditRepo{db: db},
	}
}

var (
	_ store.OrgRepository            = (*pgOrgRepo)(nil)
	_ store.OrgNodeRepository        = (*pgOrgNodeRepo)(nil)
	_ store.BudgetRepository         = (*pgBudgetRepo)(nil)
	_ store.KeysRepository           = (*pgKeysRepo)(nil)
	_ store.ModelsRepository         = (*pgModelsRepo)(nil)
	_ store.ModelAllowlistRepository = (*pgModelAllowlistRepo)(nil)
	_ store.AuditRepository          = (*pgAuditRepo)(nil)
)
