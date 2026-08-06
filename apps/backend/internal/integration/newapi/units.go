package newapi

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
	pkgorg "github.com/tokenjoy/backend/internal/support/org"
)

var PriceFromRatio = modelcatalog.PriceFromRatio
var RatioFromPrice = modelcatalog.RatioFromPrice

func NewAPIGroupForDepartment(departmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s", pkgorg.NewAPIGroupPrefix, departmentID)
}
