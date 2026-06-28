package datasource

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type RemoteDepartment struct {
	ExternalID       string
	Name             string
	ParentExternalID string
	LeaderUserID     string
}

type RemoteMember struct {
	ExternalID           string
	Name                 string
	Email                string
	Mobile               string
	DepartmentExternalID string
	EmployeeNo           string
}

type Provider interface {
	TestConnection(ctx context.Context) error
	SearchMember(ctx context.Context, keyword string) (RemoteMember, error)
	ListDepartments(ctx context.Context) ([]RemoteDepartment, error)
	ListMembers(ctx context.Context) ([]RemoteMember, []types.ImportFailure, error)
}

type Factory interface {
	ForPlatform(platform types.Platform, cred types.Credential) (Provider, error)
}
