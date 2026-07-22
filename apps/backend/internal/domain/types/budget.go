package types

import "github.com/google/uuid"

type BudgetNode struct {
	ID              uuid.UUID    `json:"id"`
	Name            string       `json:"name"`
	ParentID        *uuid.UUID   `json:"parentId"`
	Budget          int64        `json:"budget"`
	Consumed        int64        `json:"consumed"`
	ReservedPool    *int64       `json:"reservedPool,omitempty"`
	Children        []BudgetNode `json:"children,omitempty"`
	Period          string       `json:"period"`
	MemberAvgBudget int64        `json:"memberAvgBudget"`
}

type Project struct {
	ID                uuid.UUID           `json:"id"`
	Name              string              `json:"name"`
	Budget            int64               `json:"budget"`
	Consumed          int64               `json:"consumed"`
	MemberIDs         []uuid.UUID         `json:"memberIds"`
	MemberBudgets     map[uuid.UUID]int64 `json:"memberBudgets,omitempty"`
	OwnerDepartmentID uuid.UUID           `json:"ownerDepartmentId"`
}

type UpdateProjectInput struct {
	Name              *string              `json:"name"`
	Budget            *int64               `json:"budget"`
	MemberIDs         *[]uuid.UUID         `json:"memberIds"`
	MemberBudgets     *map[uuid.UUID]int64 `json:"memberBudgets"`
	OwnerDepartmentID *uuid.UUID           `json:"ownerDepartmentId"`
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
	PersonalBudget int64     `json:"personalBudget"`
	Allocated      int64     `json:"allocated"`
	Consumed       int64     `json:"consumed"`
}

type UpdateMemberBudgetInput struct {
	PersonalBudget int64 `json:"personalBudget"`
}
