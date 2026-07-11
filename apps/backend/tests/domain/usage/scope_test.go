package usage_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestValidateWindowDayLimit(t *testing.T) {
	t.Parallel()
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(400 * 24 * time.Hour)
	err := domainusage.ValidateWindow(start, end, types.UsageGranularityDay)
	if err == nil {
		t.Fatal("expected day window validation error")
	}
}

func TestValidateGroupBy(t *testing.T) {
	t.Parallel()
	if err := domainusage.ValidateGroupBy(types.UsageGroupByModel); err != nil {
		t.Fatalf("expected valid groupBy: %v", err)
	}
	if err := domainusage.ValidateGroupBy("invalid"); err == nil {
		t.Fatal("expected invalid groupBy error")
	}
}

func TestValidateSeriesPointLimit(t *testing.T) {
	t.Parallel()
	if err := domainusage.ValidateSeriesPointLimit(types.UsageMaxSeriesPoints); err != nil {
		t.Fatalf("expected limit at max to pass: %v", err)
	}
	if err := domainusage.ValidateSeriesPointLimit(types.UsageMaxSeriesPoints + 1); err == nil {
		t.Fatal("expected too many points error")
	}
}

func TestResolveScopeDepartmentsForbidden(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: "dept-1", Name: "Root", Children: []types.Department{
			{ID: "dept-8", Name: "Admin"},
			{ID: "dept-3", Name: "Backend"},
		}},
	}
	_, err := domainusage.ResolveScopeDepartments(departments, domainusage.SessionScope{
		MemberID: "m-scoped", DepartmentID: "dept-8", Permissions: []string{permission.BudgetRead},
	}, "dept-3", domainusage.DashboardScopeConfig{})
	if err == nil {
		t.Fatal("expected forbidden for out-of-scope department")
	}
}

func TestResolveScopeDepartmentsEmptyConfigScopesToSubtree(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: "dept-1", Name: "Root", Children: []types.Department{
			{ID: "dept-8", Name: "Admin"},
			{ID: "dept-3", Name: "Backend"},
		}},
	}
	got, err := domainusage.ResolveScopeDepartments(departments, domainusage.SessionScope{
		MemberID: "m-scoped", DepartmentID: "dept-8", Permissions: []string{permission.DashboardCost},
	}, "", domainusage.DashboardScopeConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "dept-8" {
		t.Fatalf("expected subtree scope [dept-8], got %v", got)
	}
}
