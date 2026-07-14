package exchange

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

// Point ↔ display helpers using DefaultPointsPerUnit.
// For charge/lot settle paths, use company.BillingCurrency PPU via ToPointsAt / ToDisplayAt.
// Settled ledger rows already store DisplayAmount + BillingCurrency — do not reconvert.

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

// Format is a currency-symbol-free display amount for user-facing messages.
func Format(points float64) string {
	return fmt.Sprintf("%.0f", ToDisplay(points))
}
