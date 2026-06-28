package usage

import "github.com/tokenjoy/backend/internal/domain/types"

const (
	DefaultTimezone = types.UsageDefaultTimezone

	GranularityDay    = types.UsageGranularityDay
	GranularityHour   = types.UsageGranularityHour
	GranularityMinute = types.UsageGranularityMinute
	GranularityWeek   = types.UsageGranularityWeek
	GranularityMonth  = types.UsageGranularityMonth

	GroupByNone       = types.UsageGroupByNone
	GroupByDepartment = types.UsageGroupByDepartment
	GroupByMember     = types.UsageGroupByMember
	GroupByModel      = types.UsageGroupByModel

	SourceBuckets = types.UsageSourceBuckets
	SourceLogs    = types.UsageSourceLogs

	MappingAsOfIngestTime = types.UsageMappingAsOfIngestTime
	MappingAsOfQueryTime  = types.UsageMappingAsOfQueryTime

	MaxDayWindow    = types.UsageMaxDayWindow
	MaxHourWindow   = types.UsageMaxHourWindow
	MaxMinuteWindow = types.UsageMaxMinuteWindow

	MaxSeriesPoints = types.UsageMaxSeriesPoints
	MaxLogPages     = types.UsageMaxLogPages
	MaxLogEntries   = types.UsageMaxLogEntries
	LogPageSize     = types.UsageLogPageSize

	MinuteCacheTTL       = types.UsageMinuteCacheTTL
	NewAPILogsTimeout    = types.UsageNewAPILogsTimeout
	MinuteRetryAfterSecs = types.UsageMinuteRetryAfterSecs
)
