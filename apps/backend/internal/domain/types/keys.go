package types

type ProviderKey struct {
	ID             string   `json:"id"`
	Provider       string   `json:"provider"`
	Name           string   `json:"name"`
	KeyPrefix      string   `json:"keyPrefix"`
	Status         string   `json:"status"`
	Balance        *float64 `json:"balance"`
	LastUsed       *string  `json:"lastUsed"`
	CreatedAt      string   `json:"createdAt"`
	RotateEnabled  bool     `json:"rotateEnabled"`
	SecretKey      string   `json:"-"`
	RelayChannelID int      `json:"-"`
}

type PlatformKey struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	KeyPrefix       string  `json:"keyPrefix"`
	FullKey         *string `json:"fullKey,omitempty"`
	Type            string  `json:"type"`
	MemberID        *string `json:"memberId"`
	MemberName      *string `json:"memberName"`
	ProjectName     *string `json:"projectName"`
	DepartmentID    string  `json:"departmentId"`
	DepartmentName  string  `json:"departmentName"`
	BudgetGroupID   *string `json:"budgetGroupId"`
	BudgetGroupName *string `json:"budgetGroupName"`
	Status          string  `json:"status"`
	Quota           float64 `json:"quota"`
	Used            float64 `json:"used"`
	ModelWhitelist  []int64 `json:"modelWhitelist"`
	CreatedAt       string  `json:"createdAt"`
	ExpiresAt       *string `json:"expiresAt"`
}

type PlatformKeyListFilter struct {
	MemberID      string
	BudgetGroupID string
	DepartmentID  string
	Type          string
}

type KeyApproval struct {
	ID              string  `json:"id"`
	Type            string  `json:"type"`
	Applicant       string  `json:"applicant"`
	ApplicantID     string  `json:"applicantId"`
	Department      string  `json:"department"`
	Reason          string  `json:"reason"`
	RequestedQuota  float64 `json:"requestedQuota"`
	RequestedModels []int64 `json:"requestedModels"`
	Status          string  `json:"status"`
	Approver        *string `json:"approver"`
	RejectReason    *string `json:"rejectReason,omitempty"`
	CreatedAt       string  `json:"createdAt"`
	ResolvedAt      *string `json:"resolvedAt"`
}

type MemberQuotaSummary struct {
	TotalQuota   float64 `json:"totalQuota"`
	Used         float64 `json:"used"`
	Remaining    float64 `json:"remaining"`
	ReservedPool float64 `json:"reservedPool"`
}

type ApprovalQuotaCheck struct {
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
	Name           string  `json:"name"`
	MemberID       *string `json:"memberId"`
	BudgetGroupID  *string `json:"budgetGroupId"`
	Quota          float64 `json:"quota"`
	ModelWhitelist []int64 `json:"modelWhitelist"`
}

type UpdatePlatformKeyInput struct {
	Name           *string  `json:"name"`
	Quota          *float64 `json:"quota"`
	ModelWhitelist []int64  `json:"modelWhitelist"`
}

type TogglePlatformKeyInput struct {
	Enabled bool `json:"enabled"`
}

type CreateApprovalInput struct {
	Type            string  `json:"type"`
	Reason          string  `json:"reason"`
	RequestedQuota  float64 `json:"requestedQuota"`
	RequestedModels []int64 `json:"requestedModels"`
	MemberID        string  `json:"memberId"`
}

type RejectApprovalInput struct {
	Reason *string `json:"reason"`
}
