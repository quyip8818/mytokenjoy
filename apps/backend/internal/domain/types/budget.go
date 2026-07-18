package types

import "github.com/google/uuid"

type BudgetNode struct {
	ID              uuid.UUID    `json:"id"`
	Name            string       `json:"name"`
	ParentID        *uuid.UUID   `json:"parentId"`
	Budget          float64      `json:"budget"`
	Consumed        float64      `json:"consumed"`
	ReservedPool    *float64     `json:"reservedPool,omitempty"`
	Children        []BudgetNode `json:"children,omitempty"`
	Period          string       `json:"period"`
	MemberAvgBudget float64      `json:"memberAvgBudget"`
}

type Project struct {
	ID                uuid.UUID             `json:"id"`
	Name              string                `json:"name"`
	Budget            float64               `json:"budget"`
	Consumed          float64               `json:"consumed"`
	MemberIDs         []uuid.UUID           `json:"memberIds"`
	MemberBudgets     map[uuid.UUID]float64 `json:"memberBudgets,omitempty"`
	OwnerDepartmentID uuid.UUID             `json:"ownerDepartmentId"`
}

type UpdateProjectInput struct {
	Name              *string                `json:"name"`
	Budget            *float64               `json:"budget"`
	MemberIDs         *[]uuid.UUID           `json:"memberIds"`
	MemberBudgets     *map[uuid.UUID]float64 `json:"memberBudgets"`
	OwnerDepartmentID *uuid.UUID             `json:"ownerDepartmentId"`
}

type OverrunPolicyConfig struct {
	Thresholds   []int  `json:"thresholds"`
	NotifyEmail  bool   `json:"notifyEmail"`
	NotifyPhone  bool   `json:"notifyPhone"`
	NotifyIm     bool   `json:"notifyIm"`
	BlockMessage string `json:"blockMessage"`
}

type AlertRule struct {
	ID            uuid.UUID   `json:"id"`
	NodeID        uuid.UUID   `json:"nodeId"`
	NodeName      string      `json:"nodeName"`
	Thresholds    []int       `json:"thresholds"`
	NotifyRoleIDs []uuid.UUID `json:"notifyRoleIds"`
	Enabled       bool        `json:"enabled"`
}

type MemberBudget struct {
	MemberID       uuid.UUID `json:"memberId"`
	MemberName     string    `json:"memberName"`
	DepartmentID   uuid.UUID `json:"departmentId"`
	PersonalBudget float64   `json:"personalBudget"`
	Allocated      float64   `json:"allocated"`
	Consumed       float64   `json:"consumed"`
}

type UpdateMemberBudgetInput struct {
	PersonalBudget float64 `json:"personalBudget"`
}

type BudgetApproval struct {
	ID             uuid.UUID `json:"id"`
	ApplicantID    uuid.UUID `json:"-"`
	DepartmentID   uuid.UUID `json:"-"`
	ApplicantName  string    `json:"applicantName"`
	DepartmentName string    `json:"departmentName"`
	Amount         float64   `json:"amount"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	CreatedAt      string    `json:"createdAt"`
	ResolvedAt     *string   `json:"resolvedAt,omitempty"`
	RejectReason   *string   `json:"rejectReason,omitempty"`
}

type ResolveBudgetApprovalInput struct {
	Status       string  `json:"status"`
	RejectReason *string `json:"rejectReason,omitempty"`
}
