package types

import "time"

var (
	UsageLifetimeStart = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	UsageLifetimeEnd   = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	UsageDefaultTimezone = "Asia/Shanghai"

	UsageGranularityDay    = "day"
	UsageGranularityHour   = "hour"
	UsageGranularityMinute = "minute"
	UsageGranularityWeek   = "week"
	UsageGranularityMonth  = "month"

	UsageGroupByNone       = "none"
	UsageGroupByDepartment = "department"
	UsageGroupByMember     = "member"
	UsageGroupByModel      = "model"

	UsageSourceBuckets = "buckets"
	UsageSourceLedger  = "ledger"

	UsageMappingAsOfIngestTime = "ingest_time"
	UsageMappingAsOfQueryTime  = "query_time"

	UsageMaxDayWindow    = 365 * 24 * time.Hour
	UsageMaxHourWindow   = 90 * 24 * time.Hour
	UsageMaxMinuteWindow = 3 * time.Hour

	UsageMaxSeriesPoints = 10000
	UsageMaxLogPages     = 50
	UsageMaxLogEntries   = 5000
	UsageLogPageSize     = 100

	UsageMinuteCacheTTL       = 60 * time.Second
	UsageNewAPILogsTimeout    = 10 * time.Second
	UsageMinuteRetryAfterSecs = 30

	NotificationChannelLog     = "log"
	NotificationChannelWebhook = "webhook"

	NotificationStatusSent   = "sent"
	NotificationStatusFailed = "failed"

	NotificationEventSyncThreshold  = "sync_threshold_exceeded"
	NotificationEventOverrunBlocked = "overrun_blocked"
)

type UsageBucketRow struct {
	BucketStart  time.Time
	DepartmentID string
	MemberID     string
	Model        string
	Cost         float64
	CallCount    int
	InputTokens  int64
	OutputTokens int64
}

type UsageSeriesQuery struct {
	Granularity  string
	Start        time.Time
	End          time.Time
	GroupBy      string
	DepartmentID string
	MemberID     string
	Timezone     string
	ScopeDeptIDs []string
}

type UsageSeriesPoint struct {
	Bucket       string  `json:"bucket"`
	DepartmentID string  `json:"departmentId,omitempty"`
	MemberID     string  `json:"memberId,omitempty"`
	Model        string  `json:"model,omitempty"`
	Cost         float64 `json:"cost"`
	CallCount    int     `json:"callCount"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
}

type UsageSeriesResponse struct {
	Granularity   string             `json:"granularity"`
	Source        string             `json:"source"`
	Timezone      string             `json:"timezone"`
	Approximate   bool               `json:"approximate"`
	MappingAsOf   string             `json:"mappingAsOf"`
	UnmappedCount *int               `json:"unmappedCount,omitempty"`
	Truncated     *bool              `json:"truncated,omitempty"`
	Points        []UsageSeriesPoint `json:"points"`
}

type UsageAggregateQuery struct {
	Start         time.Time
	End           time.Time
	Granularity   string
	Timezone      string
	GroupBy       string
	DepartmentID  string
	DepartmentIDs []string
	MemberID      string
	ParentDeptID  string
	Limit         int
	ScopeDeptIDs  []string
}

type UsageAggregateRow struct {
	Bucket       string
	DepartmentID string
	MemberID     string
	Model        string
	Cost         float64
	CallCount    int
	InputTokens  int64
	OutputTokens int64
}

type UsageSummaryTotals struct {
	Cost         float64
	CallCount    int
	InputTokens  int64
	OutputTokens int64
}

type NotificationLogEntry struct {
	ID        string
	Channel   string
	EventType string
	Recipient string
	Payload   []byte
	Status    string
	Error     string
}

type Notification struct {
	EventType string
	Recipient string
	Payload   map[string]any
}
