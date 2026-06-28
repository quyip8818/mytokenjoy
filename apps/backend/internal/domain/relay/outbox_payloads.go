package relay

type CreateTokenOutboxPayload struct {
	PlatformKeyID string `json:"platformKeyId"`
}

type UpdateTokenOutboxPayload struct {
	PlatformKeyID string `json:"platformKeyId"`
}

type UpsertChannelOutboxPayload struct {
	ProviderKeyID string `json:"providerKeyId"`
}

type UpdateModelLimitsOutboxPayload struct {
	DepartmentID string `json:"departmentId"`
}

type RebalanceAxisOutboxPayload struct {
	AxisKind string `json:"axisKind"`
	AxisID   string `json:"axisId"`
}
