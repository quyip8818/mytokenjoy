package types

import "github.com/google/uuid"

// CreateMemberInput is the request body for POST /org/members.
// Splits user identity fields from member org fields.
type CreateMemberInput struct {
	User   CreateMemberUserInput `json:"user"`
	Member CreateMemberData      `json:"member"`
}

type CreateMemberUserInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type CreateMemberData struct {
	Alias        string    `json:"alias"`
	DepartmentID uuid.UUID `json:"departmentId"`
	EmployeeID   string    `json:"employeeId"`
	JobTitle     string    `json:"jobTitle"`
	HireDate     string    `json:"hireDate"`
}

// UpdateMemberInput is the request body for PUT /org/members/:id.
// Only member-owned fields.
type UpdateMemberInput struct {
	Alias        string    `json:"alias"`
	DepartmentID uuid.UUID `json:"departmentId"`
	EmployeeID   string    `json:"employeeId"`
	JobTitle     string    `json:"jobTitle"`
	HireDate     string    `json:"hireDate"`
	Roles        []string  `json:"roles"`
	Status       string    `json:"status"`
}

// UpdateMemberUserInput is the request body for PUT /org/members/:id/user.
// Admin updates user identity fields for a member.
type UpdateMemberUserInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}
