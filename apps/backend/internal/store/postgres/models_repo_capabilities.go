package postgres

import (
	"context"

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

func (r *pgModelsRepo) withCapabilities(ctx context.Context, items []types.ModelInfo) ([]types.ModelInfo, error) {
	if len(items) == 0 {
		return items, nil
	}
	ids := make([]int64, len(items))
	for i, item := range items {
		ids[i] = item.ModelID
	}
	byID, err := r.loadCapabilitiesBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Capabilities = byID[items[i].ModelID]
	}
	return items, nil
}

func (r *pgModelsRepo) modelByCompanyAndType(ctx context.Context, companyID int64, modelType string) (*types.ModelInfo, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models WHERE company_id = $1 AND type = $2
		ORDER BY CASE WHEN provider = $3 THEN 0 ELSE 1 END, provider
		LIMIT 1
	`, companyID, modelType, types.ProviderCustom)
	return r.scanModelWithCapabilities(ctx, row)
}

func (r *pgModelsRepo) modelByCompanyProviderAndType(ctx context.Context, companyID int64, provider, modelType string) (*types.ModelInfo, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models WHERE company_id = $1 AND provider = $2 AND type = $3
	`, companyID, provider, modelType)
	return r.scanModelWithCapabilities(ctx, row)
}

func (r *pgModelsRepo) scanModelWithCapabilities(ctx context.Context, row scannable) (*types.ModelInfo, error) {
	item, err := scanModelQueryRow(row)
	if err != nil || item == nil {
		return item, err
	}
	caps, err := r.loadCapabilities(ctx, item.ModelID)
	if err != nil {
		return nil, err
	}
	item.Capabilities = caps
	return item, nil
}

func (r *pgModelsRepo) loadCapabilitiesBatch(ctx context.Context, modelIDs []int64) (map[int64][]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT model_id, capability FROM model_capabilities
		WHERE model_id = ANY($1)
		ORDER BY model_id, capability
	`, modelIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[int64][]string, len(modelIDs))
	for rows.Next() {
		var modelID int64
		var capability string
		if err := rows.Scan(&modelID, &capability); err != nil {
			return nil, err
		}
		byID[modelID] = append(byID[modelID], capability)
	}
	return byID, rows.Err()
}

func (r *pgModelsRepo) loadCapabilities(ctx context.Context, modelID int64) ([]string, error) {
	byID, err := r.loadCapabilitiesBatch(ctx, []int64{modelID})
	if err != nil {
		return nil, err
	}
	return byID[modelID], nil
}

func (r *pgModelsRepo) replaceCapabilities(ctx context.Context, modelID int64, capabilities []string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM model_capabilities WHERE model_id = $1`, modelID); err != nil {
		return err
	}
	for _, capability := range capabilities {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO model_capabilities (model_id, capability) VALUES ($1, $2)
		`, modelID, capability); err != nil {
			return err
		}
	}
	return nil
}

func scanModelRow(rows pgx.Rows) (types.ModelInfo, error) {
	var item types.ModelInfo
	err := rows.Scan(
		&item.ModelID, &item.CompanyID, &item.Provider, &item.Type, &item.Name,
		&item.Description, &item.Endpoint,
		&item.ApiKey, &item.EndpointModelName,
		&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.MaxTokens, &item.Enabled,
	)
	return item, err
}

func scanModelQueryRow(row scannable) (*types.ModelInfo, error) {
	var item types.ModelInfo
	err := row.Scan(
		&item.ModelID, &item.CompanyID, &item.Provider, &item.Type, &item.Name,
		&item.Description, &item.Endpoint,
		&item.ApiKey, &item.EndpointModelName,
		&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.MaxTokens, &item.Enabled,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
