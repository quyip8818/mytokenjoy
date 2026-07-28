package datasource

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource/dingtalk"
)

type dingtalkProvider struct {
	client *dingtalk.Client
}

func newDingtalkProvider(client *dingtalk.Client) Provider {
	return &dingtalkProvider{client: client}
}

func (p *dingtalkProvider) TestConnection(ctx context.Context) error {
	return p.client.TestConnection(ctx)
}

func (p *dingtalkProvider) SearchMember(_ context.Context, _ string) (RemoteMember, error) {
	return RemoteMember{}, nil // ponytail: not implemented yet
}

func (p *dingtalkProvider) ListDepartments(ctx context.Context) ([]RemoteDepartment, error) {
	depts, err := p.client.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RemoteDepartment, 0, len(depts))
	for _, d := range depts {
		result = append(result, RemoteDepartment{
			ExternalID:       d.ExternalID,
			Name:             d.Name,
			ParentExternalID: d.ParentExternalID,
		})
	}
	return result, nil
}

func (p *dingtalkProvider) ListMembers(ctx context.Context) ([]RemoteMember, []types.ImportFailure, error) {
	members, err := p.client.ListMembers(ctx)
	if err != nil {
		return nil, nil, err
	}
	result := make([]RemoteMember, 0, len(members))
	for _, m := range members {
		result = append(result, RemoteMember{
			ExternalID:           m.ExternalID,
			Name:                 m.Name,
			Email:                m.Email,
			Mobile:               m.Mobile,
			DepartmentExternalID: m.DepartmentExternalID,
			EmployeeNo:           m.EmployeeNo,
		})
	}
	return result, nil, nil
}

var _ Provider = (*dingtalkProvider)(nil)
