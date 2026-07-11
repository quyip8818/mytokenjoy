//go:build testhook

package testutil

import (
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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

func PendingRebalanceCount(st store.Store, companyID int64) int {
	return pendingJobCount(st, jobs.KindRebalance, companyID)
}

func PendingOverrunCount(st store.Store, companyID int64) int {
	return pendingJobCount(st, jobs.KindOverrun, companyID)
}

func PendingWalletSyncCount(st store.Store, companyID int64) int {
	return pendingJobCount(st, jobs.KindWalletSync, companyID)
}

func pendingJobCount(st store.Store, kind string, companyID int64) int {
	ctx := CtxForCompany(companyID)
	pool := postgres.MainPool(st)
	if pool == nil {
		return 0
	}
	var count int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND (args->>'company_id')::bigint = $2
	`, kind, companyID).Scan(&count); err != nil {
		return 0
	}
	return count
}
