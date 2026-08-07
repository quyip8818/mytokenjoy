package postgres

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type redemptionRepo struct {
	db dbQuerier
}

func newRedemptionRepo(db dbQuerier) *redemptionRepo {
	return &redemptionRepo{db: db}
}

func (r *redemptionRepo) BatchInsert(ctx context.Context, codes []store.RedemptionCode) error {
	if len(codes) == 0 {
		return nil
	}
	// Build a batch INSERT statement.
	sql := `INSERT INTO redemption_codes (id, code, batch_name, face_value, currency, status, expires_at, created_by, note, created_at) VALUES `
	args := make([]any, 0, len(codes)*10)
	for i, c := range codes {
		if i > 0 {
			sql += ", "
		}
		base := i * 10
		sql += "($" + strconv.Itoa(base+1) + ",$" + strconv.Itoa(base+2) + ",$" + strconv.Itoa(base+3) +
			",$" + strconv.Itoa(base+4) + ",$" + strconv.Itoa(base+5) + ",$" + strconv.Itoa(base+6) +
			",$" + strconv.Itoa(base+7) + ",$" + strconv.Itoa(base+8) + ",$" + strconv.Itoa(base+9) +
			",$" + strconv.Itoa(base+10) + ")"
		args = append(args, c.ID, c.Code, c.BatchName, c.FaceValue, c.Currency, c.Status, c.ExpiresAt, c.CreatedBy, c.Note, c.CreatedAt)
	}
	_, err := r.db.Exec(ctx, sql, args...)
	return err
}

func (r *redemptionRepo) GetCodeForUpdate(ctx context.Context, code string) (*store.RedemptionCode, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, code, batch_name, face_value, currency, status,
			redeemed_by_company, redeemed_by_member, redeemed_at, recharge_order_id,
			expires_at, created_by, note, created_at
		FROM redemption_codes
		WHERE code = $1
		FOR UPDATE
	`, code)
	return scanRedemptionCode(row)
}

func (r *redemptionRepo) MarkUsed(ctx context.Context, id uuid.UUID, companyID, memberID, orderID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		UPDATE redemption_codes
		SET status = $1, redeemed_by_company = $2, redeemed_by_member = $3,
			redeemed_at = $4, recharge_order_id = $5
		WHERE id = $6
	`, store.RedemptionStatusUsed, companyID, memberID, now, orderID, id)
	return err
}

func (r *redemptionRepo) List(ctx context.Context, filter store.RedemptionListFilter) (store.RedemptionListResult, error) {
	// Build WHERE clause.
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if filter.BatchName != nil && *filter.BatchName != "" {
		where += " AND batch_name = $" + strconv.Itoa(argIdx)
		args = append(args, *filter.BatchName)
		argIdx++
	}
	if filter.Status != nil && *filter.Status != "" {
		where += " AND status = $" + strconv.Itoa(argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	// Count total.
	var total int
	countSQL := "SELECT COUNT(*) FROM redemption_codes " + where
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return store.RedemptionListResult{}, err
	}

	// Paginate.
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataSQL := `SELECT id, code, batch_name, face_value, currency, status,
		redeemed_by_company, redeemed_by_member, redeemed_at, recharge_order_id,
		expires_at, created_by, note, created_at
		FROM redemption_codes ` + where + ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return store.RedemptionListResult{}, err
	}
	defer rows.Close()

	var items []store.RedemptionCode
	for rows.Next() {
		rc, err := scanRedemptionCode(rows)
		if err != nil {
			return store.RedemptionListResult{}, err
		}
		items = append(items, *rc)
	}
	if err := rows.Err(); err != nil {
		return store.RedemptionListResult{}, err
	}

	return store.RedemptionListResult{Items: items, Total: total}, nil
}

func scanRedemptionCode(s scannable) (*store.RedemptionCode, error) {
	var rc store.RedemptionCode
	err := s.Scan(
		&rc.ID, &rc.Code, &rc.BatchName, &rc.FaceValue, &rc.Currency, &rc.Status,
		&rc.RedeemedByCompany, &rc.RedeemedByMember, &rc.RedeemedAt, &rc.RechargeOrderID,
		&rc.ExpiresAt, &rc.CreatedBy, &rc.Note, &rc.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rc, nil
}
