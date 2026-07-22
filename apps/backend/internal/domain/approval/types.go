package approval

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// CreateInput is the request to create a new approval.
type CreateInput struct {
	Type           types.ApprovalType
	ApplicantID    uuid.UUID
	ApplicantName  string
	DepartmentID   uuid.UUID
	DepartmentName string
	Metadata       json.RawMessage
}

// ApproverInfo identifies who approved/rejected.
type ApproverInfo struct {
	ID   uuid.UUID
	Name string
}
