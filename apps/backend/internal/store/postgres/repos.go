package postgres

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func newDomainRepoSet(ctx context.Context, db dbQuerier) domainRepos {
	return domainRepos{
		org:    &pgOrgRepo{ctx: ctx, db: db},
		budget: &pgBudgetRepo{ctx: ctx, db: db},
		keys:   &pgKeysRepo{ctx: ctx, db: db},
		models: &pgModelsRepo{ctx: ctx, db: db},
		audit:  &pgAuditRepo{ctx: ctx, db: db},
	}
}

var (
	_ store.OrgRepository    = (*pgOrgRepo)(nil)
	_ store.BudgetRepository = (*pgBudgetRepo)(nil)
	_ store.KeysRepository   = (*pgKeysRepo)(nil)
	_ store.ModelsRepository = (*pgModelsRepo)(nil)
	_ store.AuditRepository  = (*pgAuditRepo)(nil)
)
