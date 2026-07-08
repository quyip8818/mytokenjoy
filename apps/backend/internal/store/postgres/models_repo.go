package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelsRepo struct {
	db        dbQuerier
	allowlist *pgModelAllowlistRepo
	catalog   modelCatalog
}

func (r *pgModelsRepo) Allowlist() store.ModelAllowlistRepository {
	return r.allowlist
}

func (r *pgModelsRepo) Models(ctx context.Context) ([]types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	items, err := r.queryModels(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE company_id = $1 OR company_id = $2
		ORDER BY CASE WHEN company_id = $1 THEN 0 ELSE 1 END, model_id
	`, r.catalog.globalCompanyID(), companyID)
	if err != nil {
		return nil, err
	}
	return r.withCapabilities(ctx, modelcatalog.DedupeEffective(items))
}

func (r *pgModelsRepo) ModelByType(ctx context.Context, modelType string) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	item, err := r.modelByCompanyAndType(ctx, companyID, modelType)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	return r.modelByCompanyAndType(ctx, r.catalog.globalCompanyID(), modelType)
}

func (r *pgModelsRepo) ModelByProviderType(ctx context.Context, provider, modelType string) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	item, err := r.modelByCompanyProviderAndType(ctx, companyID, provider, modelType)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	return r.modelByCompanyProviderAndType(ctx, r.catalog.globalCompanyID(), provider, modelType)
}

func (r *pgModelsRepo) ModelByID(ctx context.Context, modelID int64) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE model_id = $1 AND (company_id = $2 OR company_id = $3)
	`, modelID, r.catalog.globalCompanyID(), companyID)
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

func (r *pgModelsRepo) ModelByIDs(ctx context.Context, modelIDs []int64) ([]types.ModelInfo, error) {
	if len(modelIDs) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	items, err := r.queryModels(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE model_id = ANY($1) AND (company_id = $2 OR company_id = $3)
		ORDER BY model_id
	`, modelIDs, r.catalog.globalCompanyID(), companyID)
	if err != nil {
		return nil, err
	}
	return r.withCapabilities(ctx, items)
}

func (r *pgModelsRepo) InsertModel(ctx context.Context, model types.ModelInfo) (types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	if model.Visibility == "" {
		return types.ModelInfo{}, fmt.Errorf("model visibility is required")
	}
	var modelID int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO models (
			company_id, provider, type, name, description, visibility, endpoint,
			input_price, output_price, max_context, enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING model_id
	`, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Visibility, model.Endpoint,
		model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled).Scan(&modelID)
	if err != nil {
		return types.ModelInfo{}, fmt.Errorf("insert model: %w", err)
	}
	if err := r.replaceCapabilities(ctx, modelID, model.Capabilities); err != nil {
		return types.ModelInfo{}, err
	}
	model.ModelID = modelID
	model.CompanyID = companyID
	return model, nil
}

func (r *pgModelsRepo) UpdateModel(ctx context.Context, model types.ModelInfo) error {
	companyID := store.CompanyID(ctx)
	tag, err := r.db.Exec(ctx, `
		UPDATE models SET
			provider = $3,
			type = $4,
			name = $5,
			description = $6,
			visibility = $7,
			endpoint = $8,
			input_price = $9,
			output_price = $10,
			max_context = $11,
			enabled = $12,
			updated_at = NOW()
		WHERE model_id = $1 AND company_id = $2
	`, model.ModelID, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Visibility, model.Endpoint,
		model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled)
	if err != nil {
		return fmt.Errorf("update model %d: %w", model.ModelID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("model %d not found in tenant scope", model.ModelID)
	}
	return r.replaceCapabilities(ctx, model.ModelID, model.Capabilities)
}

func (r *pgModelsRepo) DeleteModel(ctx context.Context, modelID int64) error {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, `
		UPDATE org_nodes SET default_model_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND default_model_id = $2
	`, companyID, modelID); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE org_nodes SET fallback_model_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND fallback_model_id = $2
	`, companyID, modelID); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `
		DELETE FROM models WHERE model_id = $1 AND company_id = $2
	`, modelID, companyID)
	if err != nil {
		return fmt.Errorf("delete model %d: %w", modelID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("model %d not found in tenant scope", modelID)
	}
	return nil
}

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
		&item.Description, &item.Visibility, &item.Endpoint,
		&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.Enabled,
	)
	return item, err
}

func scanModelQueryRow(row scannable) (*types.ModelInfo, error) {
	var item types.ModelInfo
	err := row.Scan(
		&item.ModelID, &item.CompanyID, &item.Provider, &item.Type, &item.Name,
		&item.Description, &item.Visibility, &item.Endpoint,
		&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.Enabled,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
