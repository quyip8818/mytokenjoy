package company

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// AuditAppender is the minimal store surface for appending operation logs.
type AuditAppender interface {
	Audit() store.AuditRepository
}

func AppendPlatformOperationLog(ctx context.Context, st AuditAppender, companyID int64, action, operatorID, target, detail string) error {
	companyCtx := WithContext(ctx, Context{CompanyID: companyID})
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
