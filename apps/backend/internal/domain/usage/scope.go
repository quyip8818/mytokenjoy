package usage

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

type SessionScope struct {
	MemberID     string
	DepartmentID string
	Permissions  []string
}

func ResolveTimezone(timezone string) string {
	if timezone != "" {
		return timezone
	}
	return types.UsageDefaultTimezone
}

func ResolveScopeDepartments(
	departments []types.Department,
	scope SessionScope,
	requestedDeptID string,
) ([]string, error) {
	if requestedDeptID != "" {
		if !IsDepartmentAccessible(departments, scope, requestedDeptID) {
			return nil, domain.Forbidden("Department not accessible")
		}
		return nil, nil
	}
	if hasOrgWideDashboardAccess(scope.Permissions) {
		return nil, nil
	}
	return collectSubtreeIDs(departments, scope.DepartmentID), nil
}

func IsDepartmentAccessible(departments []types.Department, scope SessionScope, targetDeptID string) bool {
	if hasOrgWideDashboardAccess(scope.Permissions) {
		return org.FindDepartment(departments, targetDeptID) != nil
	}
	allowed := collectSubtreeIDs(departments, scope.DepartmentID)
	for _, id := range allowed {
		if id == targetDeptID {
			return true
		}
	}
	return false
}

func hasOrgWideDashboardAccess(permissions []string) bool {
	if authz.HasAny(permissions, "*") {
		return true
	}
	return authz.HasAny(permissions, permission.DashboardCost, permission.DashboardUsage)
}

func collectSubtreeIDs(departments []types.Department, rootID string) []string {
	root := org.FindDepartment(departments, rootID)
	if root == nil {
		return []string{rootID}
	}
	ids := []string{root.ID}
	for _, child := range root.Children {
		ids = append(ids, collectSubtreeIDsFromNode(child)...)
	}
	return ids
}

func collectSubtreeIDsFromNode(dept types.Department) []string {
	ids := []string{dept.ID}
	for _, child := range dept.Children {
		ids = append(ids, collectSubtreeIDsFromNode(child)...)
	}
	return ids
}

func ParseSeriesTimeRange(startRaw, endRaw, granularity, timezone string) (time.Time, time.Time, error) {
	loc, err := common.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, err := parseBoundary(startRaw, loc, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseBoundary(endRaw, loc, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, domain.BadRequest("end must be after start")
	}
	if err := ValidateWindow(start, end, granularity); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start.UTC(), end.UTC(), nil
}

func parseBoundary(value string, loc *time.Location, isEnd bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, domain.BadRequest("start and end are required")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.In(loc), nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, domain.BadRequest(fmt.Sprintf("invalid time: %s", value))
	}
	if isEnd {
		return common.TruncateInTZ(t, 24*time.Hour, loc).Add(24 * time.Hour), nil
	}
	return common.TruncateInTZ(t, 24*time.Hour, loc), nil
}

func ValidateWindow(start, end time.Time, granularity string) error {
	duration := end.Sub(start)
	switch granularity {
	case types.UsageGranularityDay:
		if duration > types.UsageMaxDayWindow {
			return domain.Validation("day query window exceeds 365 days")
		}
	case types.UsageGranularityHour:
		if duration > types.UsageMaxHourWindow {
			return domain.Validation("hour query window exceeds 90 days")
		}
	case types.UsageGranularityMinute:
		if duration > types.UsageMaxMinuteWindow {
			return domain.TooManyRequests("minute query window exceeds 3 hours")
		}
	default:
		return domain.BadRequest("invalid granularity")
	}
	return nil
}

func ValidateGroupBy(groupBy string) error {
	switch groupBy {
	case "", types.UsageGroupByNone, types.UsageGroupByDepartment, types.UsageGroupByMember, types.UsageGroupByModel:
		return nil
	default:
		return domain.BadRequest("invalid groupBy")
	}
}

func ValidateSeriesPointLimit(count int) error {
	if count > types.UsageMaxSeriesPoints {
		return domain.Validation("too many points in response")
	}
	return nil
}
