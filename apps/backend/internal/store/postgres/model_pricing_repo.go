package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type modelPricingRepo struct {
	db dbQuerier
}

func (r *modelPricingRepo) CurrentPrice(ctx context.Context, companyID uuid.UUID, modelType string, at time.Time) (*store.ModelPricingRow, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, company_id, model_type, input_price, output_price, effective_from, note, created_at
		FROM model_pricing
		WHERE company_id = $1 AND model_type = $2 AND effective_from <= $3
		ORDER BY effective_from DESC
		LIMIT 1
	`, companyID, modelType, at)

	var p store.ModelPricingRow
	err := row.Scan(&p.ID, &p.CompanyID, &p.ModelType, &p.InputPrice, &p.OutputPrice, &p.EffectiveFrom, &p.Note, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *modelPricingRepo) CurrentPricesBatch(ctx context.Context, companyID uuid.UUID, at time.Time) ([]store.ModelPricingRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (model_type)
			id, company_id, model_type, input_price, output_price, effective_from, note, created_at
		FROM model_pricing
		WHERE company_id = $1 AND effective_from <= $2
		ORDER BY model_type, effective_from DESC
	`, companyID, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []store.ModelPricingRow
	for rows.Next() {
		var p store.ModelPricingRow
		if err := rows.Scan(&p.ID, &p.CompanyID, &p.ModelType, &p.InputPrice, &p.OutputPrice, &p.EffectiveFrom, &p.Note, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *modelPricingRepo) History(ctx context.Context, companyID uuid.UUID, modelType string) ([]store.ModelPricingRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, model_type, input_price, output_price, effective_from, note, created_at
		FROM model_pricing
		WHERE company_id = $1 AND model_type = $2
		ORDER BY effective_from DESC
	`, companyID, modelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []store.ModelPricingRow
	for rows.Next() {
		var p store.ModelPricingRow
		if err := rows.Scan(&p.ID, &p.CompanyID, &p.ModelType, &p.InputPrice, &p.OutputPrice, &p.EffectiveFrom, &p.Note, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *modelPricingRepo) Insert(ctx context.Context, row store.ModelPricingRow) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.Must(uuid.NewV7())
	}
	if row.EffectiveFrom.IsZero() {
		row.EffectiveFrom = time.Now()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO model_pricing (id, company_id, model_type, input_price, output_price, effective_from, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (company_id, model_type, effective_from) DO NOTHING
	`, row.ID, row.CompanyID, row.ModelType, row.InputPrice, row.OutputPrice, row.EffectiveFrom, row.Note)
	return err
}
