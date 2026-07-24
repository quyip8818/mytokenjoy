package types

import "github.com/google/uuid"

type AuditSettings struct {
	ContentRetentionEnabled bool `json:"contentRetentionEnabled"`
}

// AuditMeta carries operator context for audit logging.
// Embedded in input structs with json:"-" — set by handler, invisible to JSON.
type AuditMeta struct {
	OperatorID   uuid.UUID `json:"-"`
	OperatorName string    `json:"-"`
	IP           string    `json:"-"`
}

type OperationLog struct {
	ID         uuid.UUID `json:"id"`
	Action     string    `json:"action"`
	Operator   string    `json:"operator"`
	OperatorID uuid.UUID `json:"operatorId"`
	ActorType  string    `json:"actorType"`
	Target     string    `json:"target"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	CreatedAt  string    `json:"createdAt"`
}

type CallLog struct {
	ID             uuid.UUID `json:"id"`
	Caller         string    `json:"caller"`
	CallerID       string    `json:"callerId"`
	CallerType     string    `json:"callerType"`
	Model          string    `json:"model"`
	Provider       string    `json:"provider"`
	InputTokens    float64   `json:"inputTokens"`
	OutputTokens   float64   `json:"outputTokens"`
	LatencyMs      float64   `json:"latencyMs"`
	Status         string    `json:"status"`
	Cost           float64   `json:"cost"`
	CreatedAt      string    `json:"createdAt"`
	PreviewSnippet string    `json:"previewSnippet"`
}

type AuditOperationsQueryParams struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Action     string `json:"action"`
	OperatorID string `json:"operatorId"`
	Keyword    string `json:"keyword"`
	From       string `json:"from"`
	To         string `json:"to"`
}

type AuditCallsQueryParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Model    string `json:"model"`
	Status   string `json:"status"`
	CallerID string `json:"callerId"`
	Keyword  string `json:"keyword"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type OperationDailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type CallsSummary struct {
	TotalCalls   int     `json:"totalCalls"`
	ErrorCount   int     `json:"errorCount"`
	ErrorRate    float64 `json:"errorRate"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
}
