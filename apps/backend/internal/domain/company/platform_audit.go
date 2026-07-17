package company

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// AuditAppender is the minimal store surface for appending operation logs.
type AuditAppender interface {
	Audit() store.AuditRepository
}

func AppendPlatformOperationLog(ctx context.Context, st AuditAppender, companyID uuid.UUID, action string, operatorID uuid.UUID, target, detail string) error {
	companyCtx := WithContext(ctx, Context{CompanyID: companyID})
	return st.Audit().AppendOperationLog(companyCtx, types.OperationLog{
		ID:         uuid.Must(uuid.NewV7()),
		Action:     action,
		Operator:   operatorID.String(),
		OperatorID: operatorID,
		ActorType:  store.ActorTypePlatform,
		Target:     target,
		Detail:     detail,
		CreatedAt:  time.Now().UTC().Format("2006-01-02 15:04"),
	})
}
