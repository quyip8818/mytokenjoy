//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestCredentialEncryptRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	key := common.DevDefaultKey()
	payload, err := json.Marshal(types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_pg", AppSecret: "secret_pg",
	})
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Credential().SaveCredential(types.PlatformFeishu, encrypted); err != nil {
		t.Fatal(err)
	}
	stored, err := st.Credential().GetCredential()
	if err != nil || stored == nil {
		t.Fatalf("expected stored credential, err=%v stored=%v", err, stored)
	}
	if stored.Platform != types.PlatformFeishu {
		t.Fatalf("unexpected platform %s", stored.Platform)
	}
}

func TestAppendSyncLogPersists(t *testing.T) {
	st := testPostgresStore(t)
	entry := types.SyncLog{
		ID: "sync-pg-1", Time: "2026-06-10 10:00",
		Type: types.SyncTypeManual, Result: types.SyncResultSuccess, Detail: "ok",
	}
	if err := st.Org().AppendSyncLog(entry); err != nil {
		t.Fatal(err)
	}
	logs := st.Org().SyncLogs()
	found := false
	for _, log := range logs {
		if log.ID == entry.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected persisted sync log")
	}
}

func TestUsageBucketQuerySeriesHour(t *testing.T) {
	st := testPostgresStore(t)
	ctx := context.Background()
	bucket := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: "dept-hour", MemberID: "m-hour",
		Model: "gpt-4o", CostCNY: 9, CallCount: 2,
	}); err != nil {
		t.Fatal(err)
	}
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].CostCNY != 9 {
		t.Fatalf("expected one hour point with cost 9, got %+v", points)
	}
}
