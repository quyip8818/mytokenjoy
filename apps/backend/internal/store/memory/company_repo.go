package memory

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryCompanyRepo struct {
	store *Store
}

func (r *memoryCompanyRepo) GetByID(ctx context.Context, id int64) (*store.Company, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	t, ok := r.store.companies[id]
	if !ok {
		return nil, nil
	}
	copy := t
	return &copy, nil
}

func (r *memoryCompanyRepo) GetBySlug(ctx context.Context, slug string) (*store.Company, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, t := range r.store.companies {
		if t.Slug == slug {
			copy := t
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryCompanyRepo) Create(ctx context.Context, company store.Company) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now().UTC()
	}
	if company.UpdatedAt.IsZero() {
		company.UpdatedAt = company.CreatedAt
	}
	r.store.companies[company.ID] = company
	if _, ok := r.store.dataByCompany[company.ID]; !ok {
		r.store.dataByCompany[company.ID] = store.Snapshot{}
	}
	return nil
}

func (r *memoryCompanyRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	t, ok := r.store.companies[id]
	if !ok {
		return nil
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	r.store.companies[id] = t
	return nil
}

func (r *memoryCompanyRepo) UpdatePackageID(ctx context.Context, id int64, packageID *string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	t, ok := r.store.companies[id]
	if !ok {
		return nil
	}
	t.PackageID = packageID
	t.UpdatedAt = time.Now().UTC()
	r.store.companies[id] = t
	return nil
}

func (r *memoryCompanyRepo) UpdateWalletAccountID(ctx context.Context, id int64, walletAccountID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	t, ok := r.store.companies[id]
	if !ok {
		return nil
	}
	t.NewAPIWalletAccountID = &walletAccountID
	t.UpdatedAt = time.Now().UTC()
	r.store.companies[id] = t
	return nil
}

func (r *memoryCompanyRepo) UpdateRootDeptID(ctx context.Context, id int64, rootDeptID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	t, ok := r.store.companies[id]
	if !ok {
		return nil
	}
	t.RootDeptID = &rootDeptID
	t.UpdatedAt = time.Now().UTC()
	r.store.companies[id] = t
	return nil
}

func (r *memoryCompanyRepo) List(ctx context.Context) ([]store.Company, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	out := make([]store.Company, 0, len(r.store.companies))
	for _, t := range r.store.companies {
		out = append(out, t)
	}
	return out, nil
}

var _ store.CompanyRepository = (*memoryCompanyRepo)(nil)
