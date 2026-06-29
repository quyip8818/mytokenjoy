package org

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
)

type syncDiff struct {
	addDepartments    []datasource.RemoteDepartment
	updateDepartments []datasource.RemoteDepartment
	removeDepartment  []types.Department
	addMembers        []datasource.RemoteMember
	updateMembers     []datasource.RemoteMember
	removeMembers     []types.Member
}

func buildSyncDiff(
	localDepts []types.Department,
	localMembers []types.Member,
	remoteDepts []datasource.RemoteDepartment,
	remoteMembers []datasource.RemoteMember,
) syncDiff {
	remoteDeptMap := make(map[string]datasource.RemoteDepartment, len(remoteDepts))
	for _, dept := range remoteDepts {
		remoteDeptMap[dept.ExternalID] = dept
	}
	remoteMemberMap := make(map[string]datasource.RemoteMember, len(remoteMembers))
	for _, member := range remoteMembers {
		remoteMemberMap[member.ExternalID] = member
	}

	diff := syncDiff{}
	localImportedDepts := make(map[string]types.Department)
	for _, dept := range localDepts {
		if dept.ExternalID == nil || isManualDeptSource(dept.Source) {
			continue
		}
		localImportedDepts[*dept.ExternalID] = dept
		remote, ok := remoteDeptMap[*dept.ExternalID]
		if !ok {
			diff.removeDepartment = append(diff.removeDepartment, dept)
			continue
		}
		if remote.Name != dept.Name {
			diff.updateDepartments = append(diff.updateDepartments, remote)
		}
	}
	for _, remote := range remoteDepts {
		if _, ok := localImportedDepts[remote.ExternalID]; !ok {
			diff.addDepartments = append(diff.addDepartments, remote)
		}
	}

	for _, member := range localMembers {
		if member.ExternalID == nil || isManualMemberSource(member.Source) {
			continue
		}
		remote, ok := remoteMemberMap[*member.ExternalID]
		if !ok {
			diff.removeMembers = append(diff.removeMembers, member)
			continue
		}
		if remote.Name != member.Name || remote.Email != member.Email || remote.Mobile != member.Phone {
			diff.updateMembers = append(diff.updateMembers, remote)
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
			diff.addMembers = append(diff.addMembers, remote)
		}
	}
	return diff
}
