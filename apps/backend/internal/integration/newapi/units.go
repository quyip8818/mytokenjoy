package newapi

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

// PriceFromRatio delegates to modelcatalog.PriceFromRatio for backward compatibility.
var PriceFromRatio = modelcatalog.PriceFromRatio

// RatioFromPrice delegates to modelcatalog.RatioFromPrice for backward compatibility.
var RatioFromPrice = modelcatalog.RatioFromPrice

func NewAPIGroupForDepartment(departmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s", common.NewAPIGroupPrefix, departmentID)
}
