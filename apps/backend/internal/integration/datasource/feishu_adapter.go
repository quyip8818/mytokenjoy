package datasource

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource/feishu"
)

type feishuProvider struct {
	client *feishu.Client
}

func newFeishuProvider(client *feishu.Client) Provider {
	return &feishuProvider{client: client}
}

func (p *feishuProvider) TestConnection(ctx context.Context) error {
	return p.client.TestConnection(ctx)
}

func (p *feishuProvider) SearchMember(ctx context.Context, keyword string) (RemoteMember, error) {
	member, err := p.client.SearchMember(ctx, keyword)
	if err != nil {
		return RemoteMember{}, err
	}
	return mapFeishuMember(member), nil
}

func (p *feishuProvider) ListDepartments(ctx context.Context) ([]RemoteDepartment, error) {
	departments, err := p.client.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RemoteDepartment, 0, len(departments))
	for _, dept := range departments {
		result = append(result, mapFeishuDepartment(dept))
	}
	return result, nil
}

func (p *feishuProvider) ListMembers(ctx context.Context) ([]RemoteMember, []types.ImportFailure, error) {
	members, failures, err := p.client.ListMembers(ctx)
	if err != nil {
		return nil, nil, err
	}
	result := make([]RemoteMember, 0, len(members))
	for _, member := range members {
		result = append(result, mapFeishuMember(member))
	}
	return result, failures, nil
}

func mapFeishuDepartment(dept feishu.Department) RemoteDepartment {
	return RemoteDepartment{
		ExternalID:       dept.ExternalID,
		Name:             dept.Name,
		ParentExternalID: dept.ParentExternalID,
		LeaderUserID:     dept.LeaderUserID,
	}
}

func mapFeishuMember(member feishu.Member) RemoteMember {
	return RemoteMember{
		ExternalID:           member.ExternalID,
		Name:                 member.Name,
		Email:                member.Email,
		Mobile:               member.Mobile,
		DepartmentExternalID: member.DepartmentExternalID,
		EmployeeNo:           member.EmployeeNo,
	}
}

var _ Provider = (*feishuProvider)(nil)
