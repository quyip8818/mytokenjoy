package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type companyRepo struct {
	db dbQuerier
}

func newCompanyRepo(db dbQuerier) *companyRepo {
	return &companyRepo{db: db}
}

func (r *companyRepo) GetByID(ctx context.Context, id uuid.UUID) (*store.Company, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, type, status, root_dept_id, newapi_wallet_user_id, newapi_wallet_username, authz_revision,
			billing_currency, fifo_head_lot_id, wallet_remain,
			created_at, updated_at
		FROM companies WHERE id = $1
	`, id)
	return scanCompanyExtendedOptional(row)
}

func (r *companyRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]store.Company, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, name, type, status, root_dept_id, newapi_wallet_user_id, newapi_wallet_username, authz_revision,
			billing_currency, fifo_head_lot_id, wallet_remain,
			created_at, updated_at
		FROM companies WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var companies []store.Company
	for rows.Next() {
		c, err := scanCompanyExtended(rows)
		if err != nil {
			return nil, err
		}
		companies = append(companies, *c)
	}
	return companies, rows.Err()
}

func (r *companyRepo) Create(ctx context.Context, company store.Company) error {
	if company.BillingCurrency == "" {
		company.BillingCurrency = common.ResolveBillingCurrency("")
	}
	if company.Type == "" {
		company.Type = store.CompanyTypeSelfhosted
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO companies (
			id, name, type, status, root_dept_id, newapi_wallet_user_id, newapi_wallet_username, authz_revision,
			billing_currency, fifo_head_lot_id, wallet_remain,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, company.ID, company.Name, company.Type, company.Status,
		company.RootDeptID, company.NewAPIWalletCompanyID, company.NewAPIWalletUsername, company.AuthzRevision,
		company.BillingCurrency, company.FIFOHeadLotID, company.WalletRemain,
		company.CreatedAt, company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create company: %w", err)
	}
	return nil
}

func (r *companyRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE companies SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *companyRepo) UpdateNewAPIWalletCompanyID(ctx context.Context, id uuid.UUID, walletCompanyID int64) error {
	if walletCompanyID <= 0 {
		return fmt.Errorf("newapi wallet company id must be positive")
	}
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET newapi_wallet_user_id = $2, updated_at = NOW() WHERE id = $1
	`, id, walletCompanyID)
	return err
}

func (r *companyRepo) UpdateRootDeptID(ctx context.Context, id uuid.UUID, rootDeptID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET root_dept_id = $2, updated_at = NOW() WHERE id = $1
	`, id, rootDeptID)
	return err
}

func (r *companyRepo) List(ctx context.Context) ([]store.Company, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, type, status, root_dept_id, newapi_wallet_user_id, newapi_wallet_username, authz_revision,
			billing_currency, fifo_head_lot_id, wallet_remain,
			created_at, updated_at
		FROM companies ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var companies []store.Company
	for rows.Next() {
		t, err := scanCompanyExtended(rows)
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

func (r *companyRepo) GetAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error) {
	var revision int64
	err := r.db.QueryRow(ctx, `SELECT authz_revision FROM companies WHERE id = $1`, id).Scan(&revision)
	return revision, err
}

func (r *companyRepo) BumpAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error) {
	var revision int64
	err := r.db.QueryRow(ctx, `
		UPDATE companies SET authz_revision = authz_revision + 1, updated_at = NOW()
		WHERE id = $1
		RETURNING authz_revision
	`, id).Scan(&revision)
	return revision, err
}

func (r *companyRepo) LockForUpdate(ctx context.Context, id uuid.UUID) (*store.Company, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, type, status, root_dept_id, newapi_wallet_user_id, newapi_wallet_username, authz_revision,
			billing_currency, fifo_head_lot_id, wallet_remain,
			created_at, updated_at
		FROM companies WHERE id = $1 FOR UPDATE
	`, id)
	return scanCompanyExtended(row)
}

func (r *companyRepo) ApplyWalletDelta(ctx context.Context, id uuid.UUID, delta int64, fifoHeadLotID *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET
			wallet_remain = wallet_remain + $2,
			fifo_head_lot_id = COALESCE($3, fifo_head_lot_id),
			updated_at = NOW()
		WHERE id = $1
	`, id, delta, fifoHeadLotID)
	return err
}

func (r *companyRepo) SetWalletRemain(ctx context.Context, id uuid.UUID, walletRemain int64, fifoHeadLotID *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE companies SET
			wallet_remain = $2,
			fifo_head_lot_id = COALESCE($3, fifo_head_lot_id),
			updated_at = NOW()
		WHERE id = $1
	`, id, walletRemain, fifoHeadLotID)
	return err
}

func scanCompanyExtendedOptional(row scannable) (*store.Company, error) {
	c, err := scanCompanyExtended(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func scanCompanyExtended(row scannable) (*store.Company, error) {
	var c store.Company
	err := row.Scan(&c.ID, &c.Name, &c.Type, &c.Status,
		&c.RootDeptID, &c.NewAPIWalletCompanyID, &c.NewAPIWalletUsername, &c.AuthzRevision,
		&c.BillingCurrency, &c.FIFOHeadLotID, &c.WalletRemain,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

var _ store.CompanyRepository = (*companyRepo)(nil)
