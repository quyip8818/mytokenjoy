package org

import (
	"fmt"
	"time"
)

func (s *service) GetDepartmentTree() []Department {
	return s.store.Org().Departments()
}

func (s *service) CreateDepartment(name, parentID string) Department {
	return Department{
		ID:   fmt.Sprintf("dept-%d", time.Now().UnixMilli()),
		Name: name, ParentID: &parentID, MemberCount: 0,
	}
}

func (s *service) UpdateDepartment(id, name string) Department {
	return Department{
		ID: id, Name: name, ParentID: nil, MemberCount: 0,
	}
}

func (s *service) DeleteDepartment(id string) error {
	_ = id
	return nil
}
