package smssync

import (
	"context"

	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/store"
)

// RepoModelStore adapts store.ModelsRepository to the ModelStore interface.
type RepoModelStore struct {
	repo store.ModelsRepository
}

func NewRepoModelStore(repo store.ModelsRepository) *RepoModelStore {
	return &RepoModelStore{repo: repo}
}

func (s *RepoModelStore) UpsertFromSMS(ctx context.Context, model sms.CatalogModel) error {
	return s.repo.UpsertFromSMS(ctx, model.ModelID, model.DisplayName, model.Provider, model.CallType)
}
