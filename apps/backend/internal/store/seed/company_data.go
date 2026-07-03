package seed

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

func DefaultCompany(cfg config.Config) store.Company {
	return store.Company{
		ID:     DefaultCompanyID,
		Slug:   config.DefaultCompanySlug,
		Name:   cfg.ResolvedCompanyName(),
		Status: store.CompanyStatusActive,
	}
}
