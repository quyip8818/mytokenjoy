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
	Budget         int64       `json:"budget"`
	Consumed       int64       `json:"consumed"`
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

type MemberBudgetSummary struct {
	TotalBudget  int64 `json:"totalBudget"`
	Consumed     int64 `json:"consumed"`
	Remaining    int64 `json:"remaining"`
	ReservedPool int64 `json:"reservedPool"`
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
	Budget         int64       `json:"budget"`
	ModelWhitelist []uuid.UUID `json:"modelWhitelist"`

	AuditMeta `json:"-"`
}

type UpdatePlatformKeyInput struct {
	Name           *string     `json:"name"`
	Budget         *int64      `json:"budget"`
	ModelWhitelist []uuid.UUID `json:"modelWhitelist"`
}

type TogglePlatformKeyInput struct {
	Enabled bool `json:"enabled"`
}
