package exchange

import "github.com/tokenjoy/backend/internal/pkg/common"

// Point ↔ display helpers. Default PPU = common.DefaultPointsPerUnit (CNY).
// Use for budget/key/wallet point amounts only.
// Settled call fees already store display on ledger.DisplayAmount — do not reconvert.

func ToPoints(display float64) float64 {
	return ToPointsAt(display, common.DefaultPointsPerUnit)
}

func ToDisplay(points float64) float64 {
	return ToDisplayAt(points, common.DefaultPointsPerUnit)
}

func ToPointsAt(display float64, pointsPerUnit int64) float64 {
	if pointsPerUnit <= 0 {
		return 0
	}
	return display * float64(pointsPerUnit)
}

func ToDisplayAt(points float64, pointsPerUnit int64) float64 {
	if pointsPerUnit <= 0 {
		return 0
	}
	return points / float64(pointsPerUnit)
}
