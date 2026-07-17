package authz

import (
	"context"

	"github.com/google/uuid"
)

type RevisionReader interface {
	GetAuthzRevision(ctx context.Context, companyID uuid.UUID) (int64, error)
}
