package seed

import "github.com/tokenjoy/backend/internal/store"

func DefaultCompany() store.Company {
	return store.Company{
		ID:     DefaultCompanyID,
		Slug:   "default",
		Name:   "Default Company",
		Status: store.CompanyStatusActive,
	}
}
