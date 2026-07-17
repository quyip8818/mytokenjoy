package usage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
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
	dept1 := uuid.MustParse("00000000-0000-7000-0000-00000000dd01")
	dept8 := uuid.MustParse("00000000-0000-7000-0000-00000000dd08")
	dept3 := uuid.MustParse("00000000-0000-7000-0000-00000000dd03")
	departments := []types.Department{
		{ID: dept1, Name: "Root", Children: []types.Department{
			{ID: dept8, Name: "Admin"},
			{ID: dept3, Name: "Backend"},
		}},
	}
	_, err := domainusage.ResolveScopeDepartments(departments, domainusage.SessionScope{
		MemberID: uuid.MustParse("00000000-0000-7000-0000-00000000ee01"), DepartmentID: dept8, Permissions: []string{permission.BudgetRead},
	}, dept3, domainusage.DashboardScopeConfig{})
	if err == nil {
		t.Fatal("expected forbidden for out-of-scope department")
	}
}

func TestResolveScopeDepartmentsEmptyConfigScopesToSubtree(t *testing.T) {
	t.Parallel()
	dept1 := uuid.MustParse("00000000-0000-7000-0000-00000000dd01")
	dept8 := uuid.MustParse("00000000-0000-7000-0000-00000000dd08")
	dept3 := uuid.MustParse("00000000-0000-7000-0000-00000000dd03")
	departments := []types.Department{
		{ID: dept1, Name: "Root", Children: []types.Department{
			{ID: dept8, Name: "Admin"},
			{ID: dept3, Name: "Backend"},
		}},
	}
	got, err := domainusage.ResolveScopeDepartments(departments, domainusage.SessionScope{
		MemberID: uuid.MustParse("00000000-0000-7000-0000-00000000ee01"), DepartmentID: dept8, Permissions: []string{permission.DashboardCost},
	}, uuid.Nil, domainusage.DashboardScopeConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != dept8 {
		t.Fatalf("expected subtree scope [dept-8], got %v", got)
	}
}
