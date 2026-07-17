//go:build testhook

package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
)

func UsageBucketRows(st store.Store) []types.UsageBucketRow {
	ctx := Ctx()
	rows, err := st.Usage().QueryFilteredBuckets(ctx, types.UsageAggregateQuery{
		Start:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
		Timezone: "UTC",
	})
	if err != nil {
		return nil
	}
	return rows
}

func NotificationLogs(st store.Store) []types.NotificationLogEntry {
	return NotificationLogsForCompany(st, contract.DefaultCompanyID)
}

func NotificationLogsForCompany(st store.Store, companyID uuid.UUID) []types.NotificationLogEntry {
	logs, err := postgres.ListNotificationLogs(CtxForCompany(companyID), postgres.MainPool(st), companyID)
	if err != nil {
		return nil
	}
	return logs
}
