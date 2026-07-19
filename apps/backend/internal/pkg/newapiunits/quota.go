package newapiunits

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func FormatModelLimits(callTypes []string) string {
	if len(callTypes) == 0 {
		return ""
	}
	return strings.Join(callTypes, ",")
}

func EffectiveWhitelistIDs(keyWhitelist, deptAllowed []uuid.UUID) []uuid.UUID {
	if len(keyWhitelist) == 0 {
		return append([]uuid.UUID{}, deptAllowed...)
	}
	allowed := make(map[uuid.UUID]struct{}, len(deptAllowed))
	for _, id := range deptAllowed {
		allowed[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0, len(keyWhitelist))
	for _, id := range keyWhitelist {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func NewAPIGroupForDepartment(departmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s", common.NewAPIGroupPrefix, departmentID)
}
