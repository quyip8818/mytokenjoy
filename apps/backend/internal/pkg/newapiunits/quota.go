package newapiunits

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func NewAPIGroupForDepartment(departmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s", common.NewAPIGroupPrefix, departmentID)
}
