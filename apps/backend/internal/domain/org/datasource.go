package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
)

func (s *service) GetDataSourceStatus() DataSourceStatus {
	return s.store.Org().DataSourceStatus()
}

func (s *service) TestDataSource(ctx context.Context, cred Credential) (DataSourceTestResult, error) {
	if err := cred.Validate(); err != nil {
		msg := err.Error()
		return DataSourceTestResult{Success: false, Message: &msg}, nil
	}
	provider, err := s.providerForCredential(cred)
	if err != nil {
		return DataSourceTestResult{}, err
	}
	if err := provider.TestConnection(ctx); err != nil {
		msg := err.Error()
		return DataSourceTestResult{Success: false, Message: &msg}, nil
	}
	return DataSourceTestResult{Success: true}, nil
}

func (s *service) UpdateDataSource(ctx context.Context, cred Credential, force bool) error {
	if err := cred.Validate(); err != nil {
		return domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	current := s.store.Org().DataSourceStatus()
	if current.Connected && current.Platform != nil && *current.Platform != cred.Platform && !force {
		return domain.NewDomainError(
			domain.StatusUnprocessable,
			"platform change requires force=true",
		)
	}

	testResult, err := s.TestDataSource(ctx, cred)
	if err != nil {
		return err
	}
	if !testResult.Success {
		message := "invalid credential"
		if testResult.Message != nil {
			message = *testResult.Message
		}
		return domain.NewDomainError(domain.StatusUnprocessable, message)
	}

	if current.Connected && current.Platform != nil && *current.Platform != cred.Platform && force {
		if err := s.store.Credential().ClearCredential(); err != nil {
			return err
		}
	}

	if err := s.saveCredential(cred); err != nil {
		return err
	}

	platform := cred.Platform
	status := s.store.Org().DataSourceStatus()
	status.Connected = true
	status.Platform = &platform
	return s.store.Org().SetDataSourceStatus(status)
}

func (s *service) SearchDataSource(ctx context.Context, keyword string) (DataSourceSearchResult, error) {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return DataSourceSearchResult{}, nil
	}

	provider, _, err := s.providerForStored()
	if err != nil {
		return DataSourceSearchResult{}, err
	}

	remote, err := provider.SearchMember(ctx, trimmed)
	if err != nil {
		return DataSourceSearchResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	departments, err := provider.ListDepartments(ctx)
	if err != nil {
		return DataSourceSearchResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	deptName := remote.DepartmentExternalID
	for _, dept := range departments {
		if dept.ExternalID == remote.DepartmentExternalID {
			deptName = dept.Name
			break
		}
	}

	localDepts := s.store.Org().Departments()
	mappingOK := false
	for _, dept := range orgutil.FlattenDepartmentTree(localDepts) {
		if dept.ExternalID != nil && *dept.ExternalID == remote.DepartmentExternalID {
			mappingOK = true
			if path := orgutil.GetDeptPath(localDepts, dept.ID); path != nil {
				deptName = *path
			}
			break
		}
	}

	return DataSourceSearchResult{
		Name:       remote.Name,
		Department: deptName,
		MappingOK:  mappingOK,
	}, nil
}

type fixedProvider struct {
	departments []datasource.RemoteDepartment
	members     []datasource.RemoteMember
}

func (p *fixedProvider) TestConnection(ctx context.Context) error { return nil }

func (p *fixedProvider) SearchMember(ctx context.Context, keyword string) (datasource.RemoteMember, error) {
	return datasource.RemoteMember{}, fmt.Errorf("not implemented")
}

func (p *fixedProvider) ListDepartments(ctx context.Context) ([]datasource.RemoteDepartment, error) {
	return p.departments, nil
}

func (p *fixedProvider) ListMembers(ctx context.Context) ([]datasource.RemoteMember, []types.ImportFailure, error) {
	return p.members, nil, nil
}
