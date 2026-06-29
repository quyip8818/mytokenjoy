package periodutil

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/timeutil"
)

type ResolvedRange struct {
	Start       time.Time
	End         time.Time
	Granularity string
	Timezone    string
}

func Resolve(params types.CostQueryParams, now time.Time, timezone string) (ResolvedRange, error) {
	if timezone == "" {
		timezone = domainusage.DefaultTimezone
	}
	loc, err := timeutil.LoadLocation(timezone)
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
		granularity = domainusage.GranularityDay
	}

	var start, end time.Time
	switch period {
	case types.CostPeriodCurrentMonth:
		start = timeutil.TruncateMonthInTZ(nowLocal, loc)
		end = start.AddDate(0, 1, 0)
	case types.CostPeriodLastMonth:
		end = timeutil.TruncateMonthInTZ(nowLocal, loc)
		start = end.AddDate(0, -1, 0)
	case types.CostPeriodLast7Days:
		end = timeutil.TruncateInTZ(nowLocal.Add(24*time.Hour), 24*time.Hour, loc)
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
		start = timeutil.TruncateMonthInTZ(nowLocal, loc)
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
	return timeutil.TruncateInTZ(t, 24*time.Hour, loc).Add(24 * time.Hour).UTC(), nil
}
