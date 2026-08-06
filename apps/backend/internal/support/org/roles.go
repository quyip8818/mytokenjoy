package org

import (
	"github.com/tokenjoy/backend/internal/domain/types"
)

func ContainsRole(roles []string, roleName string) bool {
	for _, role := range roles {
		if role == roleName {
			return true
		}
	}
	return false
}

func CountMembersByRole(members []types.Member, roleName string) int {
	count := 0
	for _, member := range members {
		if ContainsRole(member.Roles, roleName) {
			count++
		}
	}
	return count
}
