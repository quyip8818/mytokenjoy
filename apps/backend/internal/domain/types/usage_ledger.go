package types

import "time"

const (
	EventTypeCallSettled    = "call_settled"
	SourceWebhook           = "webhook"
	SourceReconcile         = "reconcile"
	IdempotencyPrefixNewAPI = "newapi:"
	PreviewSnippetMaxLen    = 200
	CallerTypeMember        = "member"
	CallerTypePlatformKey   = "platform_key"
	CallStatusSuccess       = "success"
)

type UsageCallDetail struct {
	Caller         string  `json:"caller"`
	CallerID       string  `json:"callerId"`
	CallerType     string  `json:"callerType"`
	Provider       string  `json:"provider"`
	Status         string  `json:"status"`
	LatencyMs      float64 `json:"latencyMs"`
	PreviewSnippet string  `json:"previewSnippet"`
}

type UsageLedgerEntry struct {
	ID               string
	CompanyID        int64
	EventType        string
	IdempotencyKey   string
	Amount           float64
	LotID            string
	SegmentIndex     int
	DisplayAmount    float64
	BillingCurrency  string
	DepartmentID     string
	MemberID         *string
	ProjectID        *string
	PlatformKeyID    string
	PlatformKeyScope string
	Source           string
	OccurredAt       time.Time
	PeriodKey        string
	Model            string
	InputTokens      int64
	OutputTokens     int64
	CallDetail       UsageCallDetail
	CreatedAt        time.Time
}
