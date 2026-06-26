package types

type BudgetNode struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	ParentID     *string      `json:"parentId"`
	Budget       float64      `json:"budget"`
	Consumed     float64      `json:"consumed"`
	ReservedPool *float64     `json:"reservedPool,omitempty"`
	Children     []BudgetNode `json:"children,omitempty"`
	Period       string       `json:"period"`
}

type BudgetGroup struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Budget        float64  `json:"budget"`
	Consumed      float64  `json:"consumed"`
	MemberIDs     []string `json:"memberIds"`
	DepartmentIDs []string `json:"departmentIds"`
}

type OverrunPolicyConfig struct {
	Thresholds   []int  `json:"thresholds"`
	NotifyEmail  bool   `json:"notifyEmail"`
	NotifyPhone  bool   `json:"notifyPhone"`
	NotifyIm     bool   `json:"notifyIm"`
	BlockMessage string `json:"blockMessage"`
}

type AlertRule struct {
	ID            string   `json:"id"`
	NodeID        string   `json:"nodeId"`
	NodeName      string   `json:"nodeName"`
	Thresholds    []int    `json:"thresholds"`
	NotifyRoleIDs []string `json:"notifyRoleIds"`
	Enabled       bool     `json:"enabled"`
}

type MemberBudgetQuota struct {
	MemberID      string  `json:"memberId"`
	MemberName    string  `json:"memberName"`
	DepartmentID  string  `json:"departmentId"`
	PersonalQuota float64 `json:"personalQuota"`
	Allocated     float64 `json:"allocated"`
	Used          float64 `json:"used"`
}

type UpdateMemberQuotaInput struct {
	PersonalQuota float64 `json:"personalQuota"`
}

type MemberQuotaPool struct {
	PersonalQuota float64
}
