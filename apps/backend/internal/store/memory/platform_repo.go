package memory

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryPlatformRepo struct {
	store *Store
}

func (r *memoryPlatformRepo) GetOperatorByEmail(ctx context.Context, email string) (*store.PlatformOperator, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, op := range r.store.platformOperators {
		if op.Email == email {
			copy := op
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("operator not found")
}

func (r *memoryPlatformRepo) GetOperatorByID(ctx context.Context, id string) (*store.PlatformOperator, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	op, ok := r.store.platformOperators[id]
	if !ok {
		return nil, fmt.Errorf("operator not found")
	}
	copy := op
	return &copy, nil
}

func (r *memoryPlatformRepo) CreateOperator(ctx context.Context, op store.PlatformOperator) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.platformOperators == nil {
		r.store.platformOperators = make(map[string]store.PlatformOperator)
	}
	r.store.platformOperators[op.ID] = op
	return nil
}

func (r *memoryPlatformRepo) CountOperators(ctx context.Context) (int, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return len(r.store.platformOperators), nil
}

var _ store.PlatformRepository = (*memoryPlatformRepo)(nil)
