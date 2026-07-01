package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type inviteRepo struct {
	db dbQuerier
}

func newInviteRepo(db dbQuerier) *inviteRepo {
	return &inviteRepo{db: db}
}

func (r *inviteRepo) CreateInvite(ctx context.Context, invite store.CompanyInvite) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_invites (id, company_id, email, role, token, expires_at, accepted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invite.ID, invite.CompanyID, invite.Email, invite.Role, invite.Token,
		invite.ExpiresAt, invite.AcceptedAt, invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (r *inviteRepo) GetInviteByToken(ctx context.Context, token string) (*store.CompanyInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, email, role, token, expires_at, accepted_at, created_at
		FROM company_invites WHERE token = $1
	`, token)
	var inv store.CompanyInvite
	if err := row.Scan(&inv.ID, &inv.CompanyID, &inv.Email, &inv.Role, &inv.Token,
		&inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) MarkInviteAccepted(ctx context.Context, id string, acceptedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_invites SET accepted_at = $2 WHERE id = $1
	`, id, acceptedAt)
	return err
}

type platformRepo struct {
	db dbQuerier
}

func newPlatformRepo(db dbQuerier) *platformRepo {
	return &platformRepo{db: db}
}

func (r *platformRepo) GetOperatorByEmail(ctx context.Context, email string) (*store.PlatformOperator, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM platform_operators WHERE email = $1
	`, email)
	return scanPlatformOperator(row)
}

func (r *platformRepo) GetOperatorByID(ctx context.Context, id string) (*store.PlatformOperator, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM platform_operators WHERE id = $1
	`, id)
	return scanPlatformOperator(row)
}

func (r *platformRepo) CreateOperator(ctx context.Context, op store.PlatformOperator) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO platform_operators (id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, op.ID, op.Email, op.PasswordHash, op.Status, op.CreatedAt, op.UpdatedAt)
	return err
}

func (r *platformRepo) CountOperators(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM platform_operators`).Scan(&count)
	return count, err
}

func scanPlatformOperator(row scannable) (*store.PlatformOperator, error) {
	var op store.PlatformOperator
	if err := row.Scan(&op.ID, &op.Email, &op.PasswordHash, &op.Status, &op.CreatedAt, &op.UpdatedAt); err != nil {
		return nil, err
	}
	return &op, nil
}

type billingRepo struct {
	db dbQuerier
}

func newBillingRepo(db dbQuerier) *billingRepo {
	return &billingRepo{db: db}
}

func (r *billingRepo) CreateRechargeOrder(ctx context.Context, order store.RechargeOrder) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO company_recharge_orders (
			id, company_id, amount, source, idempotency_key, newapi_topup_ref, status, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, order.ID, order.CompanyID, order.Amount, order.Source, order.IdempotencyKey,
		order.NewAPITopupRef, order.Status, order.CreatedBy, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *billingRepo) GetRechargeOrder(ctx context.Context, id string) (*store.RechargeOrder, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, amount, source, idempotency_key, newapi_topup_ref, status, created_by, created_at, updated_at
		FROM company_recharge_orders WHERE id = $1
	`, id)
	var o store.RechargeOrder
	if err := row.Scan(&o.ID, &o.CompanyID, &o.Amount, &o.Source, &o.IdempotencyKey,
		&o.NewAPITopupRef, &o.Status, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *billingRepo) UpdateRechargeStatus(ctx context.Context, id, status string, topupRef *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE company_recharge_orders SET status = $2, newapi_topup_ref = COALESCE($3, newapi_topup_ref), updated_at = NOW()
		WHERE id = $1
	`, id, status, topupRef)
	return err
}

func (r *billingRepo) ListRechargeOrders(ctx context.Context, companyID int64) ([]store.RechargeOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, amount, source, idempotency_key, newapi_topup_ref, status, created_by, created_at, updated_at
		FROM company_recharge_orders WHERE company_id = $1 ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []store.RechargeOrder
	for rows.Next() {
		var o store.RechargeOrder
		if err := rows.Scan(&o.ID, &o.CompanyID, &o.Amount, &o.Source, &o.IdempotencyKey,
			&o.NewAPITopupRef, &o.Status, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

var (
	_ store.InviteRepository   = (*inviteRepo)(nil)
	_ store.PlatformRepository = (*platformRepo)(nil)
	_ store.BillingRepository  = (*billingRepo)(nil)
)
