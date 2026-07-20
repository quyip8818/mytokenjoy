package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

// ForEachActiveCompany runs fn for every active, non-testing company with a company-scoped context.
func ForEachActiveCompany(ctx context.Context, companies store.CompanyRepository, fn func(context.Context, store.Company) error) error {
	list, err := companies.List(ctx)
	if err != nil {
		return err
	}
	for _, co := range list {
		if co.Status != store.CompanyStatusActive {
			continue
		}
		if co.Type == store.CompanyTypeTesting {
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
		Type:      co.Type,
		Status:    co.Status,
	}
	if id, ok := store.ConfiguredNewAPIWalletCompanyID(&co); ok {
		info.NewAPIWalletCompanyID = id
	}
	return info
}

// ResolveNewAPIWalletCompanyID returns the NewAPI wallet company ID from context (fast path)
// or falls back to querying the DB.
func ResolveNewAPIWalletCompanyID(ctx context.Context, companies store.CompanyRepository) (int64, bool) {
	if info, ok := FromContext(ctx); ok && info.NewAPIWalletCompanyID > 0 {
		return info.NewAPIWalletCompanyID, true
	}
	co, err := companies.GetByID(ctx, CompanyID(ctx))
	if err != nil || co == nil {
		return 0, false
	}
	return store.ConfiguredNewAPIWalletCompanyID(co)
}
