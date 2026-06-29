package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryModelsRepo struct{ store *Store }

func (r *memoryModelsRepo) Models() []types.ModelInfo {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneModels(r.store.data.Models)
}

func (r *memoryModelsRepo) SetModels(models []types.ModelInfo) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Models = store.CloneModels(models)
	return nil
}

func (r *memoryModelsRepo) RoutingRules() []types.RoutingRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoutingRules(r.store.data.RoutingRules)
}

func (r *memoryModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.RoutingRules = store.CloneRoutingRules(rules)
	return nil
}
