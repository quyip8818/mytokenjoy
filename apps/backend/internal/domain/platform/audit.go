package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
)

func AppendAudit(ctx context.Context, st store.Store, action, operatorID, target, detail string) error {
	companyCtx := ctxcompany.With(ctx, ctxcompany.Info{CompanyID: 1})
	return st.Audit().AppendOperationLog(companyCtx, types.OperationLog{
		ID:         fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Action:     action,
		Operator:   operatorID,
		OperatorID: operatorID,
		ActorType:  store.ActorTypePlatform,
		Target:     target,
		Detail:     detail,
		CreatedAt:  time.Now().UTC().Format("2006-01-02 15:04"),
	})
}
