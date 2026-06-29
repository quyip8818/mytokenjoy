package org

import "github.com/tokenjoy/backend/internal/domain/types"

func localDeptID(platform types.Platform, externalID string) string {
	return "dept-" + string(platform) + "-" + externalID
}

func localMemberID(platform types.Platform, externalID string) string {
	return "m-" + string(platform) + "-" + externalID
}

func stringPtr(value string) *string {
	return &value
}

func isManualDeptSource(source *string) bool {
	return source != nil && *source == types.DeptSourceManual
}

func isManualMemberSource(source string) bool {
	return source == types.MemberSourceManual
}
