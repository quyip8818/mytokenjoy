package snapshot

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func defaultCompany(cfg config.Config) store.Company {
	return store.Company{
		ID:     contract.DefaultCompanyID,
		Name:   cfg.ResolvedCompanyName(),
		Status: store.CompanyStatusActive,
	}
}
