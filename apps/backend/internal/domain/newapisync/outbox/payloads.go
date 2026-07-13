package outbox

type CreateKeyOutboxPayload struct {
	CompanyID     int64  `json:"companyId"`
	PlatformKeyID string `json:"platformKeyId"`
}

type UpsertChannelOutboxPayload struct {
	CompanyID     int64  `json:"companyId"`
	ProviderKeyID string `json:"providerKeyId"`
}

type UpdateModelLimitsOutboxPayload struct {
	CompanyID    int64  `json:"companyId"`
	DepartmentID string `json:"departmentId"`
}

type RebalanceAxisOutboxPayload struct {
	CompanyID int64  `json:"companyId"`
	AxisKind  string `json:"axisKind"`
	AxisID    string `json:"axisId"`
}
