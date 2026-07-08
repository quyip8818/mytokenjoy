package budget

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

const PeriodMonthly = "monthly"

type ResolvedRange struct {
	Start       time.Time
	End         time.Time
	Granularity string
	Timezone    string
}

// SnapshotKey resolves the budget_snapshots period_key for an org period spec at a point in time.
// Realtime gates (precheck, overrun) and snapshot projection use time.Now().UTC().
// Ledger rows store period_key from entry.OccurredAt via DepartmentPeriodKey.
func SnapshotKey(orgPeriod string, at time.Time) string {
	if orgPeriod != "" && orgPeriod != PeriodMonthly {
		return orgPeriod
	}
	return at.UTC().Format("2006-01")
}

func Resolve(params types.CostQueryParams, now time.Time, timezone string) (ResolvedRange, error) {
	if timezone == "" {
		timezone = types.UsageDefaultTimezone
	}
	loc, err := common.LoadLocation(timezone)
	if err != nil {
		return ResolvedRange{}, err
	}
	nowLocal := now.In(loc)
	period := types.CostPeriod(params.Period)
	if period == "" {
		period = types.CostPeriodCurrentMonth
	}
	granularity := params.Granularity
	if granularity == "" {
		granularity = types.UsageGranularityDay
	}

	var start, end time.Time
	switch period {
	case types.CostPeriodCurrentMonth:
		start = common.TruncateMonthInTZ(nowLocal, loc)
		end = start.AddDate(0, 1, 0)
	case types.CostPeriodLastMonth:
		end = common.TruncateMonthInTZ(nowLocal, loc)
		start = end.AddDate(0, -1, 0)
	case types.CostPeriodLast7Days:
		end = common.TruncateInTZ(nowLocal.Add(24*time.Hour), 24*time.Hour, loc)
		start = end.AddDate(0, 0, -7)
	case types.CostPeriodCustom:
		if params.StartDate == "" || params.EndDate == "" {
			return ResolvedRange{}, fmt.Errorf("startDate and endDate are required for custom period")
		}
		start, err = parseDateStart(params.StartDate, loc)
		if err != nil {
			return ResolvedRange{}, err
		}
		end, err = parseDateEnd(params.EndDate, loc)
		if err != nil {
			return ResolvedRange{}, err
		}
	default:
		start = common.TruncateMonthInTZ(nowLocal, loc)
		end = start.AddDate(0, 1, 0)
	}

	return ResolvedRange{
		Start:       start.UTC(),
		End:         end.UTC(),
		Granularity: granularity,
		Timezone:    timezone,
	}, nil
}

func PreviousRange(r ResolvedRange) ResolvedRange {
	duration := r.End.Sub(r.Start)
	return ResolvedRange{
		Start:       r.Start.Add(-duration),
		End:         r.Start,
		Granularity: r.Granularity,
		Timezone:    r.Timezone,
	}
}

func parseDateStart(value string, loc *time.Location) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.In(loc).UTC(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}
	return t.UTC(), nil
}

func parseDateEnd(value string, loc *time.Location) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.In(loc).UTC(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}
	return common.TruncateInTZ(t, 24*time.Hour, loc).Add(24 * time.Hour).UTC(), nil
}
