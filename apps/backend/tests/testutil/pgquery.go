//go:build testhook

package testutil

import (
	"time"

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
	logs, err := postgres.ListNotificationLogs(Ctx(), postgres.MainPool(st), contract.DefaultCompanyID)
	if err != nil {
		return nil
	}
	return logs
}
