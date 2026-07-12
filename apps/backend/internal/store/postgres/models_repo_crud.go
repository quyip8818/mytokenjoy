package postgres

import (
	"context"
	"fmt"

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
	var modelID int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO models (
			company_id, provider, type, name, description, endpoint,
			input_price, output_price, max_context, enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING model_id
	`, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
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
			endpoint = $7,
			input_price = $8,
			output_price = $9,
			max_context = $10,
			enabled = $11,
			updated_at = NOW()
		WHERE model_id = $1 AND company_id = $2
	`, model.ModelID, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
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

var _ store.ModelsRepository = (*pgModelsRepo)(nil)
