package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// ApprovalListFilter controls the approval list query.
type ApprovalListFilter struct {
	CompanyID   uuid.UUID
	Status      *types.ApprovalStatus
	Type        *types.ApprovalType
	ApplicantID *uuid.UUID
	Limit       int
	Offset      int
}

// ApprovalRepository persists unified approval requests.
type ApprovalRepository interface {
	Create(ctx context.Context, req types.ApprovalRequest) error
	Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error)
	Update(ctx context.Context, req types.ApprovalRequest) error
	List(ctx context.Context, filter ApprovalListFilter) ([]types.ApprovalRequest, int, error)
}
