package models

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) requireTenantModel(ctx context.Context, modelID uuid.UUID) (*types.ModelInfo, error) {
	model, err := s.store.Models().ModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, domain.NotFound("Not found")
	}
	if model.CompanyID == s.cfg.TokenJoyCompanyID {
		return nil, domain.Forbidden("global models are read-only")
	}
	companyID := store.CompanyID(ctx)
	if model.CompanyID != companyID {
		return nil, domain.NotFound("Not found")
	}
	return model, nil
}

func (s *service) validateModelProviderTypeAvailable(ctx context.Context, provider, modelType string) error {
	existing, err := s.store.Models().ModelByProviderType(ctx, provider, modelType)
	if err != nil {
		return err
	}
	if existing != nil {
		return domain.Validation("model already exists for provider")
	}
	return nil
}

func (s *service) validateWritableModelIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	catalog, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	if err := modelcatalog.ValidateWritableIDs(catalog, ids); err != nil {
		if errors.Is(err, modelcatalog.ErrUnknownModelID) {
			return domain.Validation("unknown model id")
		}
		if errors.Is(err, modelcatalog.ErrModelDisabled) {
			return domain.Validation("model is disabled")
		}
		return err
	}
	return nil
}
