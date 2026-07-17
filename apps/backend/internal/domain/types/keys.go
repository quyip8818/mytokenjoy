package types

import "github.com/google/uuid"

const (
	PlatformKeyScopeMember        = "member"
	PlatformKeyScopeProject       = "project"
	PlatformKeyScopeProjectMember = "project_member"
)

func ValidPlatformKeyScope(scope string) bool {
	switch scope {
	case PlatformKeyScopeMember, PlatformKeyScopeProject, PlatformKeyScopeProjectMember:
		return true
	default:
		return false
	}
}

type ProviderKey struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	Name            string    `json:"name"`
	KeyPrefix       string    `json:"keyPrefix"`
	Status          string    `json:"status"`
	CreatedAt       string    `json:"createdAt"`
	RotateEnabled   bool      `json:"rotateEnabled"`
	SecretKey       string    `json:"-"`
	NewAPIChannelID int       `json:"-"`
}

type PlatformKey struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	KeyPrefix      string      `json:"keyPrefix"`
	FullKey        *string     `json:"fullKey,omitempty"`
	Scope          string      `json:"scope"`
	MemberID       *uuid.UUID  `json:"memberId"`
	MemberName     *string     `json:"memberName"`
	ProjectID      *uuid.UUID  `json:"projectId"`
	ProjectName    *string     `json:"projectName"`
	DepartmentID   uuid.UUID   `json:"departmentId"`
	DepartmentName string      `json:"departmentName"`
	Status         string      `json:"status"`
	Budget         float64     `json:"budget"`
	Consumed       float64     `json:"consumed"`
	ModelWhitelist []uuid.UUID `json:"modelWhitelist"`
	CreatedAt      string      `json:"createdAt"`
	ExpiresAt      *string     `json:"expiresAt"`
}

type PlatformKeyListFilter struct {
	MemberID     uuid.UUID
	ProjectID    uuid.UUID
	DepartmentID uuid.UUID
	Scope        string
}

type KeyApproval struct {
	ID              uuid.UUID   `json:"id"`
	Type            string      `json:"type"`
	Applicant       string      `json:"applicant"`
	ApplicantID     uuid.UUID   `json:"applicantId"`
	Department      string      `json:"department"`
	Reason          string      `json:"reason"`
	RequestedBudget float64     `json:"requestedBudget"`
	RequestedModels []uuid.UUID `json:"requestedModels"`
	Status          string      `json:"status"`
	Approver        *string     `json:"approver"`
	RejectReason    *string     `json:"rejectReason,omitempty"`
	CreatedAt       string      `json:"createdAt"`
	ResolvedAt      *string     `json:"resolvedAt"`
}

type MemberBudgetSummary struct {
	TotalBudget  float64 `json:"totalBudget"`
	Consumed     float64 `json:"consumed"`
	Remaining    float64 `json:"remaining"`
	ReservedPool float64 `json:"reservedPool"`
}

type ApprovalBudgetCheck struct {
	Sufficient   bool    `json:"sufficient"`
	ReservedPool float64 `json:"reservedPool"`
	Requested    float64 `json:"requested"`
}

type CreateProviderKeyInput struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Key      string `json:"key"`
}

type ToggleProviderKeyInput struct {
	Enabled bool `json:"enabled"`
}

type RotateProviderKeyInput struct {
	NewKey string `json:"newKey"`
}

type CreatePlatformKeyInput struct {
	Name           string      `json:"name"`
	Scope          string      `json:"scope"`
	MemberID       *uuid.UUID  `json:"memberId"`
	ProjectID      *uuid.UUID  `json:"projectId"`
	Budget         float64     `json:"budget"`
	ModelWhitelist []uuid.UUID `json:"modelWhitelist"`
}

type UpdatePlatformKeyInput struct {
	Name           *string     `json:"name"`
	Budget         *float64    `json:"budget"`
	ModelWhitelist []uuid.UUID `json:"modelWhitelist"`
}

type TogglePlatformKeyInput struct {
	Enabled bool `json:"enabled"`
}

type CreateApprovalInput struct {
	Type            string      `json:"type"`
	Reason          string      `json:"reason"`
	RequestedBudget float64     `json:"requestedBudget"`
	RequestedModels []uuid.UUID `json:"requestedModels"`
	MemberID        uuid.UUID   `json:"memberId"`
}

type RejectApprovalInput struct {
	Reason *string `json:"reason"`
}
