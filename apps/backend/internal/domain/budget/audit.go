package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/identity/httpx"
	"github.com/tokenjoy/backend/internal/store"
)

// appendBudgetLog writes an operation log for budget changes.
// Best-effort: errors are logged, not propagated.
func (s *service) appendBudgetLog(ctx context.Context, action, target, detail string) {
	operatorName, operatorID := resolveOperator(ctx)
	_ = s.store.Audit().AppendOperationLog(ctx, types.OperationLog{
		ID:         uuid.Must(uuid.NewV7()),
		Action:     action,
		Operator:   operatorName,
		OperatorID: operatorID,
		ActorType:  store.ActorTypeMember,
		Target:     target,
		Detail:     detail,
		CreatedAt:  time.Now().UTC().Format("2006-01-02 15:04"),
	})
}

func resolveOperator(ctx context.Context) (string, uuid.UUID) {
	if session, ok := httpx.SessionFromContext(ctx); ok {
		return session.Member.Alias, session.Member.ID
	}
	return "system", uuid.Nil
}

// budgetChangeDetail formats a JSON-like detail string for budget field changes.
func budgetChangeDetail(field string, oldVal, newVal float64) string {
	return fmt.Sprintf(`{"field":%q,"old":%.2f,"new":%.2f}`, field, oldVal, newVal)
}
