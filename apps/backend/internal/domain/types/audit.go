package types

type AuditSettings struct {
	ContentRetentionEnabled bool `json:"contentRetentionEnabled"`
}

type OperationLog struct {
	ID         string `json:"id"`
	Action     string `json:"action"`
	Operator   string `json:"operator"`
	OperatorID string `json:"operatorId"`
	ActorType  string `json:"actorType"`
	Target     string `json:"target"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"createdAt"`
}

type CallLog struct {
	ID             string  `json:"id"`
	Caller         string  `json:"caller"`
	CallerID       string  `json:"callerId"`
	CallerType     string  `json:"callerType"`
	Model          string  `json:"model"`
	Provider       string  `json:"provider"`
	InputTokens    float64 `json:"inputTokens"`
	OutputTokens   float64 `json:"outputTokens"`
	LatencyMs      float64 `json:"latencyMs"`
	Status         string  `json:"status"`
	Cost           float64 `json:"cost"`
	CreatedAt      string  `json:"createdAt"`
	PreviewSnippet string  `json:"previewSnippet"`
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
