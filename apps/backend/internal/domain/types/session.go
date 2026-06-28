package types

type Member struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Phone          string   `json:"phone"`
	Email          string   `json:"email"`
	DepartmentID   string   `json:"departmentId"`
	DepartmentName string   `json:"departmentName"`
	Status         string   `json:"status"`
	Roles          []string `json:"roles"`
	Source         string   `json:"source"`
	ExternalID     *string  `json:"externalId,omitempty"`
}

type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
	MemberCount int      `json:"memberCount"`
}

type SessionContext struct {
	Member      Member   `json:"member"`
	Permissions []string `json:"permissions"`
	ReadOnly    bool     `json:"readOnly"`
}
