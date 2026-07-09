package points

import "github.com/tokenjoy/backend/internal/pkg/common"

func FromDisplay(display float64) float64 {
	return display * float64(common.DefaultPointsPerUnit)
}
