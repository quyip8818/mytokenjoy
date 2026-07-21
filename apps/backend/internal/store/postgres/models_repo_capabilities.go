package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func (r *pgModelsRepo) queryModels(ctx context.Context, query string, args ...any) ([]types.ModelInfo, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.ModelInfo, 0)
	for rows.Next() {
		item, err := scanModelRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgModelsRepo) modelByCompanyAndType(ctx context.Context, companyID uuid.UUID, modelType string) (*types.ModelInfo, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models WHERE company_id = $1 AND type = $2
		ORDER BY CASE WHEN provider = $3 THEN 0 ELSE 1 END, provider
		LIMIT 1
	`, companyID, modelType, types.ProviderCustom)
	return scanModelQueryRow(row)
}

func (r *pgModelsRepo) modelByCompanyProviderAndType(ctx context.Context, companyID uuid.UUID, provider, modelType string) (*types.ModelInfo, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models WHERE company_id = $1 AND provider = $2 AND type = $3
	`, companyID, provider, modelType)
	return scanModelQueryRow(row)
}

func scanModelRow(rows pgx.Rows) (types.ModelInfo, error) {
	var item types.ModelInfo
	err := rows.Scan(
		&item.ID, &item.CompanyID, &item.Provider, &item.Type, &item.Name,
		&item.Description, &item.Endpoint,
		&item.ApiKey, &item.EndpointModelName,
		&item.MaxContext, &item.MaxTokens, &item.Enabled,
		&item.Capabilities,
	)
	return item, err
}

func scanModelQueryRow(row scannable) (*types.ModelInfo, error) {
	var item types.ModelInfo
	err := row.Scan(
		&item.ID, &item.CompanyID, &item.Provider, &item.Type, &item.Name,
		&item.Description, &item.Endpoint,
		&item.ApiKey, &item.EndpointModelName,
		&item.MaxContext, &item.MaxTokens, &item.Enabled,
		&item.Capabilities,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
