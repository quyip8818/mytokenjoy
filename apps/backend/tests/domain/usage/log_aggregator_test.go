package usage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
	mock "github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestLogAggregatorUnmappedAndTruncated(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	stub := &mock.StubAdminClient{
		ListLogsFn: func(_ context.Context, params newapi.ListLogsParams) ([]newapi.LogEntry, error) {
			if params.Page > 1 {
				return []newapi.LogEntry{}, nil
			}
			return []newapi.LogEntry{
				{ID: 1, TokenID: 999, Quota: 100000, ModelName: "gpt-4o", CreatedAt: time.Now().Unix()},
			}, nil
		},
	}
	agg := domainusage.NewLogAggregator(stub, st, nil)
	start := time.Now().Add(-30 * time.Minute).UTC()
	end := time.Now().UTC()
	resp, err := agg.Series(context.Background(), types.UsageSeriesQuery{
		Granularity: types.UsageGranularityMinute,
		Start:       start,
		End:         end,
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLogs || !resp.Approximate {
		t.Fatalf("unexpected response meta: %+v", resp)
	}
	if resp.UnmappedCount == nil || *resp.UnmappedCount != 1 {
		t.Fatalf("expected unmappedCount=1, got %+v", resp.UnmappedCount)
	}
	if len(resp.Points) != 0 {
		t.Fatalf("expected no points for unmapped log, got %+v", resp.Points)
	}
}

func TestLogAggregatorUnavailableWithoutClient(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	agg := domainusage.NewLogAggregator(nil, st, nil)
	_, err := agg.Series(context.Background(), types.UsageSeriesQuery{
		Granularity: types.UsageGranularityMinute,
		Start:       time.Now().Add(-time.Hour).UTC(),
		End:         time.Now().UTC(),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err == nil {
		t.Fatal("expected unavailable error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain error, got %T", err)
	}
	if domainErr.RetryAfter == nil || *domainErr.RetryAfter != types.UsageMinuteRetryAfterSecs {
		t.Fatalf("expected retryAfter=%d, got %+v", types.UsageMinuteRetryAfterSecs, domainErr.RetryAfter)
	}
}

func TestLogAggregatorMinuteSuccess(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	tokenID := int64(42)
	testutil.UpsertRelayMapping(t, st, testutil.RelayMappingOpts{
		PlatformKeyID: "plk-agg", NewAPITokenID: tokenID,
	})
	createdAt := time.Now().Add(-15 * time.Minute).UTC()
	stub := &mock.StubAdminClient{
		ListLogsFn: func(_ context.Context, _ newapi.ListLogsParams) ([]newapi.LogEntry, error) {
			return []newapi.LogEntry{testutil.SampleMappedLog(tokenID, createdAt)}, nil
		},
	}
	agg := domainusage.NewLogAggregator(stub, st, nil)
	start := time.Now().Add(-30 * time.Minute).UTC()
	end := time.Now().UTC()
	resp, err := agg.Series(context.Background(), types.UsageSeriesQuery{
		Granularity: types.UsageGranularityMinute,
		Start:       start,
		End:         end,
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceLogs || !resp.Approximate || resp.MappingAsOf != types.UsageMappingAsOfQueryTime {
		t.Fatalf("unexpected response meta: %+v", resp)
	}
	if len(resp.Points) != 1 || resp.Points[0].CallCount != 1 {
		t.Fatalf("expected one minute point, got %+v", resp.Points)
	}
}

func TestLogAggregatorDoesNotWriteBuckets(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	tokenID := int64(7)
	testutil.UpsertRelayMapping(t, st, testutil.RelayMappingOpts{
		PlatformKeyID: "plk-readonly", NewAPITokenID: tokenID,
	})
	before := testutil.UsageBucketCount(st)
	stub := &mock.StubAdminClient{
		ListLogsFn: func(_ context.Context, _ newapi.ListLogsParams) ([]newapi.LogEntry, error) {
			return []newapi.LogEntry{testutil.SampleMappedLog(tokenID, time.Now().UTC())}, nil
		},
	}
	agg := domainusage.NewLogAggregator(stub, st, nil)
	_, err := agg.Series(context.Background(), types.UsageSeriesQuery{
		Granularity: types.UsageGranularityMinute,
		Start:       time.Now().Add(-30 * time.Minute).UTC(),
		End:         time.Now().UTC(),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertUsageBucketCount(t, st, before)
}
