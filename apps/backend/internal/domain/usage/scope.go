package usage

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/pkg/timeutil"
)

type SessionScope struct {
	MemberID     string
	DepartmentID string
	Permissions  []string
}

func ResolveTimezone(_ string) string {
	return DefaultTimezone
}

func ResolveScopeDepartments(
	departments []types.Department,
	scope SessionScope,
	requestedDeptID string,
) ([]string, error) {
	if requestedDeptID != "" {
		if !IsDepartmentAccessible(departments, scope, requestedDeptID) {
			return nil, domain.NewDomainError(domain.StatusForbidden, "Department not accessible")
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
		return orgutil.FindDepartment(departments, targetDeptID) != nil
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
	if permission.HasAny(permissions, "*") {
		return true
	}
	return permission.HasAny(permissions, permission.DashboardCost, permission.DashboardUsage)
}

func collectSubtreeIDs(departments []types.Department, rootID string) []string {
	root := orgutil.FindDepartment(departments, rootID)
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
	loc, err := timeutil.LoadLocation(timezone)
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
		return time.Time{}, time.Time{}, domain.NewDomainError(domain.StatusBadRequest, "end must be after start")
	}
	if err := ValidateWindow(start, end, granularity); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start.UTC(), end.UTC(), nil
}

func parseBoundary(value string, loc *time.Location, isEnd bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, domain.NewDomainError(domain.StatusBadRequest, "start and end are required")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.In(loc), nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, domain.NewDomainError(domain.StatusBadRequest, fmt.Sprintf("invalid time: %s", value))
	}
	if isEnd {
		return timeutil.TruncateInTZ(t, 24*time.Hour, loc).Add(24 * time.Hour), nil
	}
	return timeutil.TruncateInTZ(t, 24*time.Hour, loc), nil
}

func ValidateWindow(start, end time.Time, granularity string) error {
	duration := end.Sub(start)
	switch granularity {
	case GranularityDay:
		if duration > MaxDayWindow {
			return domain.NewDomainError(domain.StatusUnprocessable, "day query window exceeds 365 days")
		}
	case GranularityHour:
		if duration > MaxHourWindow {
			return domain.NewDomainError(domain.StatusUnprocessable, "hour query window exceeds 90 days")
		}
	case GranularityMinute:
		if duration > MaxMinuteWindow {
			return domain.NewDomainError(domain.StatusTooManyRequests, "minute query window exceeds 3 hours")
		}
	default:
		return domain.NewDomainError(domain.StatusBadRequest, "invalid granularity")
	}
	return nil
}

func ValidateGroupBy(groupBy string) error {
	switch groupBy {
	case "", GroupByNone, GroupByDepartment, GroupByMember, GroupByModel:
		return nil
	default:
		return domain.NewDomainError(domain.StatusBadRequest, "invalid groupBy")
	}
}

func ValidateSeriesPointLimit(count int) error {
	if count > MaxSeriesPoints {
		return domain.NewDomainError(domain.StatusUnprocessable, "too many points in response")
	}
	return nil
}
