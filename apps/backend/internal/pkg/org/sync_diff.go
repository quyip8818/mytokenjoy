package org

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

type SyncDiff struct {
	AddDepartments    []datasource.RemoteDepartment
	UpdateDepartments []datasource.RemoteDepartment
	RemoveDepartments []types.Department
	AddMembers        []datasource.RemoteMember
	UpdateMembers     []datasource.RemoteMember
	RemoveMembers     []types.Member
}

func BuildSyncDiff(
	localDepts []types.Department,
	localMembers []types.Member,
	remoteDepts []datasource.RemoteDepartment,
	remoteMembers []datasource.RemoteMember,
) SyncDiff {
	remoteDeptMap := make(map[string]datasource.RemoteDepartment, len(remoteDepts))
	for _, dept := range remoteDepts {
		remoteDeptMap[dept.ExternalID] = dept
	}
	remoteMemberMap := make(map[string]datasource.RemoteMember, len(remoteMembers))
	for _, member := range remoteMembers {
		remoteMemberMap[member.ExternalID] = member
	}

	diff := SyncDiff{}
	localImportedDepts := make(map[string]types.Department)
	for _, dept := range localDepts {
		if dept.ExternalID == nil || IsManualDeptSource(dept.Source) {
			continue
		}
		localImportedDepts[*dept.ExternalID] = dept
		remote, ok := remoteDeptMap[*dept.ExternalID]
		if !ok {
			diff.RemoveDepartments = append(diff.RemoveDepartments, dept)
			continue
		}
		if remote.Name != dept.Name {
			diff.UpdateDepartments = append(diff.UpdateDepartments, remote)
		}
	}
	for _, remote := range remoteDepts {
		if _, ok := localImportedDepts[remote.ExternalID]; !ok {
			diff.AddDepartments = append(diff.AddDepartments, remote)
		}
	}

	for _, member := range localMembers {
		if member.ExternalID == nil || IsManualMemberSource(member.Source) {
			continue
		}
		remote, ok := remoteMemberMap[*member.ExternalID]
		if !ok {
			diff.RemoveMembers = append(diff.RemoveMembers, member)
			continue
		}
		if remote.Name != member.Name || remote.Email != member.Email || remote.Mobile != member.Phone {
			diff.UpdateMembers = append(diff.UpdateMembers, remote)
		}
	}
	for _, remote := range remoteMembers {
		found := false
		for _, member := range localMembers {
			if member.ExternalID != nil && *member.ExternalID == remote.ExternalID {
				found = true
				break
			}
		}
		if !found {
			diff.AddMembers = append(diff.AddMembers, remote)
		}
	}
	return diff
}
