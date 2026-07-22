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
	ApprovalTypeKey          ApprovalType = "key"
	ApprovalTypeBudget       ApprovalType = "budget"
	ApprovalTypeMemberBudget ApprovalType = "member_budget"
)

// ApprovalRequest is the unified approval record.
type ApprovalRequest struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"-"`
	Type           ApprovalType    `json:"type"`
	Status         ApprovalStatus  `json:"status"`
	ApplicantID    uuid.UUID       `json:"applicantId"`
	ApplicantName  string          `json:"applicantName"`
	DepartmentID   uuid.UUID       `json:"departmentId,omitempty"`
	DepartmentName string          `json:"departmentName,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	ApproverID     *uuid.UUID      `json:"approverId,omitempty"`
	ApproverName   *string         `json:"approverName,omitempty"`
	RejectReason   *string         `json:"rejectReason,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	ResolvedAt     *time.Time      `json:"resolvedAt,omitempty"`
}

// --- Per-type Metadata structs ---

type KeyApprovalMeta struct {
	Reason          string      `json:"reason"`
	RequestedBudget float64     `json:"requestedBudget"`
	RequestedModels []uuid.UUID `json:"requestedModels"`
}

type BudgetApprovalMeta struct {
	Reason          string      `json:"reason"`
	RequestedBudget float64     `json:"requestedBudget"`
	RequestedModels []uuid.UUID `json:"requestedModels,omitempty"`
}

type MemberBudgetApprovalMeta struct {
	Amount int64  `json:"amount"`
	Reason string `json:"reason"`
}
