package outbox

import "github.com/google/uuid"

type CreateKeyOutboxPayload struct {
	CompanyID     uuid.UUID `json:"companyId"`
	PlatformKeyID uuid.UUID `json:"platformKeyId"`
}

type UpsertChannelOutboxPayload struct {
	CompanyID     uuid.UUID `json:"companyId"`
	ProviderKeyID string    `json:"providerKeyId"`
}

type UpdateModelLimitsOutboxPayload struct {
	CompanyID    uuid.UUID `json:"companyId"`
	DepartmentID uuid.UUID `json:"departmentId"`
}

type RebalanceAxisOutboxPayload struct {
	CompanyID uuid.UUID `json:"companyId"`
	AxisKind  string    `json:"axisKind"`
	AxisID    string    `json:"axisId"`
}
