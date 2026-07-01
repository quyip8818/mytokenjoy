package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

func CompanyID(ctx context.Context) int64 {
	return ctxcompany.ID(ctx)
}
