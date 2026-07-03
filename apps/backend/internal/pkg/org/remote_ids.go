package org

import "github.com/tokenjoy/backend/internal/domain/types"

func LocalDeptID(platform types.Platform, externalID string) string {
	return "dept-" + string(platform) + "-" + externalID
}

func LocalMemberID(platform types.Platform, externalID string) string {
	return "m-" + string(platform) + "-" + externalID
}

func IsManualDeptSource(source *string) bool {
	return source != nil && *source == types.DeptSourceManual
}

func IsManualMemberSource(source string) bool {
	return source == types.MemberSourceManual
}
