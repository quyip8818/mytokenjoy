package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ApprovalStatus string

const (
	ApprovalPending   ApprovalStatus = "pending"
	ApprovalApproved  ApprovalStatus = "approved"
	ApprovalRejected  ApprovalStatus = "rejected"
	ApprovalCancelled ApprovalStatus = "cancelled"
	ApprovalFailed    ApprovalStatus = "failed"
)

type ApprovalType string

const (
	ApprovalTypeKey                 ApprovalType = "key"
	ApprovalTypeMemberBudget        ApprovalType = "member_budget"
	ApprovalTypeProjectBudget       ApprovalType = "project_budget"
	ApprovalTypeProjectMemberBudget ApprovalType = "project_member_budget"
)

// ApprovalRequest is the unified approval record.
type ApprovalRequest struct {
	ID            uuid.UUID       `json:"id"`
	CompanyID     uuid.UUID       `json:"-"`
	Type          ApprovalType    `json:"type"`
	Status        ApprovalStatus  `json:"status"`
	ApplicantID   uuid.UUID       `json:"applicantId"`
	ApplicantName string          `json:"applicantName"`
	ScopeID       uuid.UUID       `json:"scopeId"`
	Metadata      json.RawMessage `json:"metadata"`
	ApproverID    *uuid.UUID      `json:"approverId,omitempty"`
	ApproverName  *string         `json:"approverName,omitempty"`
	RejectReason  *string         `json:"rejectReason,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	ResolvedAt    *time.Time      `json:"resolvedAt,omitempty"`
}

// --- Per-type Metadata structs ---

type KeyApprovalMeta struct {
	Reason          string      `json:"reason"`
	RequestedBudget float64     `json:"requestedBudget"`
	RequestedModels []uuid.UUID `json:"requestedModels"`
	DepartmentID    uuid.UUID   `json:"departmentId"`
	DepartmentName  string      `json:"departmentName"`
}

type MemberBudgetApprovalMeta struct {
	Amount         int64     `json:"amount"`
	Reason         string    `json:"reason"`
	DepartmentID   uuid.UUID `json:"departmentId"`
	DepartmentName string    `json:"departmentName"`
}

type ProjectBudgetApprovalMeta struct {
	ProjectID   uuid.UUID `json:"projectId"`
	ProjectName string    `json:"projectName"`
	Amount      int64     `json:"amount"`
	Reason      string    `json:"reason"`
}

type ProjectMemberBudgetApprovalMeta struct {
	ProjectID   uuid.UUID `json:"projectId"`
	ProjectName string    `json:"projectName"`
	Amount      int64     `json:"amount"`
	Reason      string    `json:"reason"`
}
