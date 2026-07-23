package remote

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

// syncMember applies remote data to an existing local member using field sync policy.
// Only fields that pass ShouldSyncField are overwritten.
func syncMember(m *types.Member, remote datasource.RemoteMember, localDept uuid.UUID, deptName string) {
	overrides := m.OverrideFields

	// immutable: employeeId — write only if local is empty
	if core.ShouldSyncField("employeeId", m.EmployeeID, overrides) {
		m.EmployeeID = remote.EmployeeNo
	}

	// user-owned: alias — skip if user has overridden
	if core.ShouldSyncField("alias", m.Alias, overrides) {
		m.Alias = remote.Name
	}

	// sync-always: department — unconditional overwrite
	m.DepartmentID = localDept
	m.DepartmentName = deptName

	// metadata always updated
	m.ExternalID = stringPtr(remote.ExternalID)
	m.Source = types.MemberSourceImported
	if m.Status == "" {
		m.Status = "active"
	}
}
