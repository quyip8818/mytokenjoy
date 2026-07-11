//go:build testhook

package testutil

import (
	"context"
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
	logs, err := postgres.ListNotificationLogs(context.Background(), postgres.MainPool(st), contract.DefaultCompanyID)
	if err != nil {
		return nil
	}
	return logs
}

func NewAPISyncOutboxEntry(st store.Store, id string) (store.AsyncJob, bool) {
	entry, found, err := postgres.GetNewAPISyncOutboxByID(context.Background(), postgres.MainPool(st), id)
	if err != nil || !found {
		return store.AsyncJob{}, false
	}
	return entry, true
}

func PendingRebalanceCount(st store.Store, companyID int64) int {
	ctx := CtxForCompany(companyID)
	entries, err := st.AsyncJobs().ClaimPendingRebalance(ctx, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}

func PendingOverrunCount(st store.Store, companyID int64) int {
	ctx := CtxForCompany(companyID)
	entries, err := st.AsyncJobs().ClaimPendingOverrun(ctx, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
