package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

type modelDiscountRepo struct {
	db dbQuerier
}

func (r *modelDiscountRepo) CurrentDiscounts(ctx context.Context, companyID uuid.UUID) ([]store.ModelDiscountRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (model_type)
			id, company_id, model_type, discount, effective_from, note, created_at
		FROM model_discount
		WHERE company_id = $1 AND effective_from <= NOW()
		ORDER BY model_type, effective_from DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []store.ModelDiscountRow
	for rows.Next() {
		var d store.ModelDiscountRow
		if err := rows.Scan(&d.ID, &d.CompanyID, &d.ModelType, &d.Discount, &d.EffectiveFrom, &d.Note, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *modelDiscountRepo) Insert(ctx context.Context, row store.ModelDiscountRow) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.Must(uuid.NewV7())
	}
	if row.EffectiveFrom.IsZero() {
		row.EffectiveFrom = time.Now()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO model_discount (id, company_id, model_type, discount, effective_from, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, row.ID, row.CompanyID, row.ModelType, row.Discount, row.EffectiveFrom, row.Note)
	return err
}

func (r *modelDiscountRepo) DeleteByCompanyAndModel(ctx context.Context, companyID uuid.UUID, modelType string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM model_discount WHERE company_id = $1 AND model_type = $2`, companyID, modelType)
	return err
}

func (r *modelDiscountRepo) DeleteAllByCompany(ctx context.Context, companyID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM model_discount WHERE company_id = $1`, companyID)
	return err
}
