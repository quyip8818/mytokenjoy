package types

import (
	"time"

	"github.com/google/uuid"
)

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
	ID               uuid.UUID
	CompanyID        uuid.UUID
	EventType        string
	IdempotencyKey   string
	Amount           int64
	LotID            uuid.UUID
	SegmentIndex     int
	DisplayAmount    float64
	BillingCurrency  string
	DepartmentID     uuid.UUID
	MemberID         *uuid.UUID
	ProjectID        *uuid.UUID
	PlatformKeyID    uuid.UUID
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
