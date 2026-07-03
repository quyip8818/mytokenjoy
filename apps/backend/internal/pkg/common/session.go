package common

import "github.com/tokenjoy/backend/internal/domain/types"

func ResolveDemoMemberName(memberID string, members []types.Member) string {
	if memberID == "" {
		return "审批人"
	}
	for _, member := range members {
		if member.ID == memberID {
			return member.Name
		}
	}
	return "审批人"
}
