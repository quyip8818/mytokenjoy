package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

// ForEachActiveCompany runs fn for every active company with a company-scoped context.
func ForEachActiveCompany(ctx context.Context, companies store.CompanyRepository, fn func(context.Context, store.Company) error) error {
	list, err := companies.List(ctx)
	if err != nil {
		return err
	}
	for _, co := range list {
		if co.Status != store.CompanyStatusActive {
			continue
		}
		entryCtx := WithContext(ctx, ContextFromStore(co))
		if err := fn(entryCtx, co); err != nil {
			return err
		}
	}
	return nil
}

// ContextFromStore builds a request Context from a companies row.
func ContextFromStore(co store.Company) Context {
	info := Context{
		CompanyID: co.ID,
		Status:    co.Status,
	}
	if id, ok := store.ConfiguredNewAPIWalletUserID(&co); ok {
		info.NewAPIWalletUserID = id
	}
	return info
}
