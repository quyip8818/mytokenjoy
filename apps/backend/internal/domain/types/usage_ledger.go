package types

import "time"

const (
	EventTypeCallSettled    = "call_settled"
	SourceWebhook           = "webhook"
	SourceCompensate        = "compensate"
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
	ID             string
	CompanyID      int64
	EventType      string
	IdempotencyKey string
	AmountCNY      float64
	DepartmentID   string
	MemberID       *string
	BudgetGroupID  *string
	PlatformKeyID  string
	Source         string
	OccurredAt     time.Time
	Model          string
	InputTokens    int64
	OutputTokens   int64
	CallDetail     UsageCallDetail
	CreatedAt      time.Time
}
