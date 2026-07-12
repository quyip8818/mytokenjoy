package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type IngestJobEnqueuer interface {
	EnqueueAfterIngest(ctx context.Context, tx store.Tx, companyID int64) error
}
