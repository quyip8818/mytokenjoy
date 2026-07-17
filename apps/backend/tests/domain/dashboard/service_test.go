package dashboard_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newDashboardSvc(t *testing.T) (dashboard.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	return dashboard.NewService(cfg, st, domainusage.NewReader(st.Usage(), st.Ledger()), domainusage.DashboardScopeConfig{
		OrgWidePermissions: []string{permission.DashboardCost, permission.DashboardUsage},
	}), st
}

func TestCostSummaryFromBuckets(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 12.5, CallCount: 3})
	summary, err := svc.CostSummary(ctx, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, uuid.Nil, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalCost != 12.5 {
		t.Fatalf("expected bucket cost 12.5, got %v", summary.TotalCost)
	}
}

func TestDailyCostsWeekGranularity(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 4})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC),
		Cost:        6,
	})
	rows, err := svc.DailyCosts(ctx, types.CostQueryParams{
		Period: string(types.CostPeriodCurrentMonth), Granularity: types.UsageGranularityWeek,
	}, uuid.Nil, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Cost != 10 {
		t.Fatalf("expected one weekly point with cost 10, got %+v", rows)
	}
}

func TestUsageSeriesRequiresParams(t *testing.T) {
	t.Parallel()
	svc, _ := newDashboardSvc(t)
	ctx := testutil.Ctx()
	_, err := svc.UsageSeries(ctx, types.UsageSeriesQuery{
		GroupBy: types.UsageGroupByNone,
	}, testutil.AdminDashboardScope())
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Status != domain.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestUsageSeriesHourFromBuckets(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 3})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
		Cost:        7,
	})
	resp, err := svc.UsageSeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != types.UsageSourceBuckets || len(resp.Points) != 2 {
		t.Fatalf("expected two hour points, got %+v", resp)
	}
}

func TestUsageTeamsConsumedFromBucketsNotSnapshot(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 18.5, CallCount: 2})
	departments, err := svc.DepartmentUsage(ctx, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, uuid.Nil, domainusage.SessionScope{
		MemberID: contract.IDMemberAdmin, Permissions: []string{permission.DashboardUsage, "*"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, dept := range departments {
		if dept.DepartmentID != contract.IDDept3 {
			continue
		}
		if dept.Consumed != 18.5 {
			t.Fatalf("expected consumed from buckets 18.5, got %v", dept.Consumed)
		}
		return
	}
	t.Fatal("dept-3 department usage not found")
}

func TestCostSummaryPeriodOverPeriod(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC),
		Cost:        5,
	})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 12.5, CallCount: 2})
	summary, err := svc.CostSummary(ctx, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, uuid.Nil, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalCost != 12.5 {
		t.Fatalf("expected current month cost 12.5, got %v", summary.TotalCost)
	}
	if summary.TotalCostMom != 150 {
		t.Fatalf("expected mom 150%%, got %v", summary.TotalCostMom)
	}
}

func TestDepartmentCostDrillDown(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{Cost: 20, CallCount: 4})
	depts, err := svc.DepartmentCosts(ctx, contract.IDDept2.String(), types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	var deptCost float64
	for _, row := range depts {
		if row.DepartmentID == contract.IDDept3 {
			deptCost = row.Cost
			break
		}
	}
	if deptCost != 20 {
		t.Fatalf("expected dept cost 20, got %v", deptCost)
	}
	members, err := svc.DepartmentMemberCosts(ctx, contract.IDDept3, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 1 || members[0].Cost != 20 {
		t.Fatalf("expected member drill-down cost 20, got %+v", members)
	}
}

func TestUsageSeriesTimezoneShanghai(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 6, 9, 16, 0, 0, 0, time.UTC),
		Cost:        3,
	})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 6, 10, 7, 0, 0, 0, time.UTC),
		Cost:        7,
	})
	resp, err := svc.UsageSeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityDay,
		Start:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		End:         time.Date(2026, 6, 11, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Points) != 1 || resp.Points[0].Cost != 10 {
		t.Fatalf("expected one shanghai day bucket with cost 10, got %+v", resp.Points)
	}
	if resp.Timezone != types.UsageDefaultTimezone {
		t.Fatalf("expected timezone %s, got %s", types.UsageDefaultTimezone, resp.Timezone)
	}
}
