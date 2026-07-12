package types

import "fmt"

const (
	DeptSourceImported = "imported"
	DeptSourceManual   = "manual"

	MemberSourceImported = "imported"
	MemberSourceManual   = "manual"
	MemberSourceInvited  = "invited"

	SyncTypeManual    = "manual"
	SyncTypeScheduled = "scheduled"
	SyncResultSuccess = "success"
	SyncResultPartial = "partial_failure"
	SyncResultFailure = "failure"

	MemberStatusActive   = "active"
	MemberStatusInactive = "inactive"
	MemberStatusPending  = "pending"
)

// OrgSyncLockName returns the per-tenant scheduler lock for org sync.
func OrgSyncLockName(companyID int64) string {
	return fmt.Sprintf("org_sync:%d", companyID)
}
