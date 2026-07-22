package approval

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// Repository is the persistence interface for approval requests.
// It mirrors store.ApprovalRepository — the Engine accepts either.
type Repository interface {
	Create(ctx context.Context, req types.ApprovalRequest) error
	Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error)
	Update(ctx context.Context, req types.ApprovalRequest) error
	List(ctx context.Context, filter store.ApprovalListFilter) ([]types.ApprovalRequest, int, error)
}
