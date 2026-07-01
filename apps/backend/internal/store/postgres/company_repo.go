package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type companyRepo struct {
	db dbQuerier
}

func newCompanyRepo(db dbQuerier) *companyRepo {
	return &companyRepo{db: db}
}

func (r *companyRepo) GetByID(ctx context.Context, id int64) (*store.Company, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, name, status, root_dept_id, newapi_wallet_account_id, package_id, created_at, updated_at
		FROM companies WHERE id = $1
	`, id)
	return scanCompany(row)
}

func (r *companyRepo) GetBySlug(ctx context.Context, slug string) (*store.Company, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, name, status, root_dept_id, newapi_wallet_account_id, package_id, created_at, updated_at
		FROM companies WHERE slug = $1
	`, slug)
	return scanCompany(row)
}

func (r *companyRepo) Create(ctx context.Context, company store.Company) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status, root_dept_id, newapi_wallet_account_id, package_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, company.ID, company.Slug, company.Name, company.Status, company.RootDeptID,
		company.NewAPIWalletAccountID, company.PackageID, company.CreatedAt, company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create company: %w", err)
	}
	return nil
}

func (r *companyRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE companies SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *companyRepo) UpdatePackageID(ctx context.Context, id int64, packageID *string) error {
	_, err := r.db.Exec(ctx, `UPDATE companies SET package_id = $2, updated_at = NOW() WHERE id = $1`, id, packageID)
	return err
}

func (r *companyRepo) UpdateWalletAccountID(ctx context.Context, id int64, walletAccountID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET newapi_wallet_account_id = $2, updated_at = NOW() WHERE id = $1
	`, id, walletAccountID)
	return err
}

func (r *companyRepo) UpdateRootDeptID(ctx context.Context, id int64, rootDeptID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET root_dept_id = $2, updated_at = NOW() WHERE id = $1
	`, id, rootDeptID)
	return err
}

func (r *companyRepo) List(ctx context.Context) ([]store.Company, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, name, status, root_dept_id, newapi_wallet_account_id, package_id, created_at, updated_at
		FROM companies ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var companies []store.Company
	for rows.Next() {
		t, err := scanCompanyRow(rows)
		if err != nil {
			return nil, err
		}
		companies = append(companies, *t)
	}
	return companies, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanCompany(row scannable) (*store.Company, error) {
	var c store.Company
	err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.Status, &c.RootDeptID,
		&c.NewAPIWalletAccountID, &c.PackageID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanCompanyRow(rows interface {
	Scan(dest ...any) error
}) (*store.Company, error) {
	return scanCompany(rows)
}

var _ store.CompanyRepository = (*companyRepo)(nil)
