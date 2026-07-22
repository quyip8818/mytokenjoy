package types

import "github.com/google/uuid"

type Member struct {
	ID             uuid.UUID `json:"id"`
	CompanyID      uuid.UUID `json:"companyId"`
	UserID         uuid.UUID `json:"userId"`
	Alias          string    `json:"alias"`
	Avatar         string    `json:"avatar,omitempty"`
	Username       string    `json:"username,omitempty"`
	EmployeeID     string    `json:"employeeId,omitempty"`
	JobTitle       string    `json:"jobTitle,omitempty"`
	HireDate       string    `json:"hireDate,omitempty"`
	DepartmentID   uuid.UUID `json:"departmentId"`
	DepartmentName string    `json:"departmentName"`
	Status         string    `json:"status"`
	Roles          []string  `json:"roles"`
	Source         string    `json:"source"`
	ExternalID     *string   `json:"externalId,omitempty"`
	PersonalBudget int64     `json:"-"`

	// Input-only fields: used for member creation/update, not persisted on members table.
	// Phone/Email are used to resolve/create the user record and stored in users table.
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Permissions []string  `json:"permissions"`
	MemberCount int       `json:"memberCount"`
}

type SessionUser struct {
	Name string `json:"name"`
}

type SessionContext struct {
	CompanyID       uuid.UUID   `json:"companyId"`
	CompanyType     string      `json:"companyType"`
	AuthzRevision   int64       `json:"authzRevision"`
	User            SessionUser `json:"user"`
	Member          Member      `json:"member"`
	Permissions     []string    `json:"permissions"`
	ReadOnly        bool        `json:"readOnly"`
	BillingCurrency string      `json:"billingCurrency"`
	QuotaPerUnit    int64       `json:"quotaPerUnit"`
}
