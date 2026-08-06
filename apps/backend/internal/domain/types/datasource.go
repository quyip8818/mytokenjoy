package types

import "context"

// RemoteDepartment represents a department fetched from an external org system (Feishu, DingTalk).
type RemoteDepartment struct {
	ExternalID       string
	Name             string
	ParentExternalID string
	LeaderUserID     string
}

// RemoteMember represents a member fetched from an external org system.
type RemoteMember struct {
	ExternalID           string
	Name                 string
	Email                string
	Mobile               string
	DepartmentExternalID string
	EmployeeNo           string
}

// DataSourceProvider fetches org data from a remote system (Feishu, DingTalk, etc.).
type DataSourceProvider interface {
	TestConnection(ctx context.Context) error
	SearchMember(ctx context.Context, keyword string) (RemoteMember, error)
	ListDepartments(ctx context.Context) ([]RemoteDepartment, error)
	ListMembers(ctx context.Context) ([]RemoteMember, []ImportFailure, error)
}

// DataSourceFactory creates a DataSourceProvider for a given platform and credential.
type DataSourceFactory interface {
	ForPlatform(platform Platform, cred Credential) (DataSourceProvider, error)
}
