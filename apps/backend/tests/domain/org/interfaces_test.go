package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/org"
)

func TestServiceImplementsSubInterfaces(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	var (
		_ org.Service           = svc
		_ org.DataSourceService = svc
		_ org.SyncService       = svc
		_ org.DepartmentService = svc
		_ org.MemberService     = svc
		_ org.RoleService       = svc
	)
}
