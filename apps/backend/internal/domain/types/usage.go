package types

import (
	"time"

	"github.com/google/uuid"
)

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
	NotificationChannelEmail   = "email"
	NotificationChannelSMS     = "sms"
	NotificationChannelInApp   = "in_app"

	NotificationEventSyncThreshold      = "sync_threshold_exceeded"
	NotificationEventOverrunBlocked     = "overrun_blocked"
	NotificationEventOverdraftExpanded  = "overdraft_expanded"
	NotificationEventBudgetAlertReached = "budget_alert_reached"
	NotificationEventKeyExpired         = "key_expired"
	NotificationEventKeyExpiringSoon    = "key_expiring_soon"
)

type UsageBucketRow struct {
	BucketStart   time.Time
	DepartmentID  uuid.UUID
	MemberID      uuid.UUID
	Model         string
	QuotaConsumed int64   // Σ ledger.quota_amount (quota)
	Cost          float64 // billing currency (Σ ledger.cost)
	CallCount     int
	InputTokens   int64
	OutputTokens  int64
}

// Spend is the billing-currency amount for dashboard / series APIs.
func (r UsageBucketRow) Spend() float64 { return r.Cost }

type UsageSeriesQuery struct {
	Granularity  string
	Start        time.Time
	End          time.Time
	GroupBy      string
	DepartmentID uuid.UUID
	MemberID     uuid.UUID
	Timezone     string
	ScopeDeptIDs []uuid.UUID
}

type UsageSeriesPoint struct {
	Bucket       string    `json:"bucket"`
	DepartmentID uuid.UUID `json:"departmentId,omitempty"`
	MemberID     uuid.UUID `json:"memberId,omitempty"`
	Model        string    `json:"model,omitempty"`
	Cost         float64   `json:"cost"`
	CallCount    int       `json:"callCount"`
	InputTokens  int64     `json:"inputTokens"`
	OutputTokens int64     `json:"outputTokens"`
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
	Start             time.Time
	End               time.Time
	Granularity       string
	Timezone          string
	GroupBy           string
	DepartmentID      uuid.UUID
	OwnerDepartmentID []uuid.UUID
	MemberID          uuid.UUID
	ParentDeptID      uuid.UUID
	Limit             int
	ScopeDeptIDs      []uuid.UUID
}

type UsageAggregateRow struct {
	Bucket        string
	DepartmentID  uuid.UUID
	MemberID      uuid.UUID
	Model         string
	QuotaConsumed float64 // Σ quota (float64 for aggregation arithmetic)
	Cost          float64 // billing currency
	CallCount     int
	InputTokens   int64
	OutputTokens  int64
}

// Spend is the billing-currency amount for dashboard / series APIs.
func (r UsageAggregateRow) Spend() float64 { return r.Cost }

type UsageSummaryTotals struct {
	QuotaConsumed float64 // Σ quota (float64 for aggregation arithmetic)
	Cost          float64 // billing currency
	CallCount     int
	InputTokens   int64
	OutputTokens  int64
}

// Spend is the billing-currency amount for dashboard / series APIs.
func (t UsageSummaryTotals) Spend() float64 { return t.Cost }

type NotificationLogEntry struct {
	ID         uuid.UUID
	Channel    string
	EventType  string
	UserID     uuid.UUID
	Title      string
	Body       string
	Payload    []byte
	SendOK     bool // true = delivered, false = failed
	Error      string
	Category   string
	GroupKey   string
	GroupCount int    // populated only for grouped queries
	Status     string // active / archived / deleted
	CreatedAt  time.Time
	ReadAt     *time.Time
	UpdatedAt  time.Time
}

// NotificationListFilter is the query filter for the inbox list endpoint.
type NotificationListFilter struct {
	UserID   uuid.UUID
	Category string
	Status   string // "unread" | "read" | ""(all)
	Archived bool
	Grouped  bool
	GroupKey string
	Cursor   *time.Time
	Limit    int
}

// NotificationListResult wraps a page of notifications with cursor info.
type NotificationListResult struct {
	Items      []NotificationLogEntry
	NextCursor *time.Time
	HasMore    bool
}

type Notification struct {
	EventType   string
	RecipientID uuid.UUID
	Payload     map[string]any
}

type NotificationPreferenceEntry struct {
	Category string
	Channel  string
	Enabled  bool
}

type NotificationLogFilter struct {
	Channel   string
	SendOK    *bool // nil = all, true = success only, false = failed only
	EventType string
	Limit     int
	Offset    int
}

type NotificationStatRow struct {
	Channel string
	SendOK  bool
	Count   int
}
