package budget

import "math"

// budgetEpsilon is the tolerance for floating-point budget comparisons.
// Values within this range of zero are treated as zero.
const budgetEpsilon = 0.001

// ClampNonNegative returns 0 if v is negative or within epsilon of zero.
func ClampNonNegative(v float64) float64 {
	if v < budgetEpsilon {
		return 0
	}
	return v
}

// BudgetExhausted returns true if consumed >= budget within floating-point tolerance.
func BudgetExhausted(consumed, budget float64) bool {
	return consumed >= budget-budgetEpsilon
}

// NearZero returns true if the absolute value is within epsilon.
func NearZero(v float64) bool {
	return math.Abs(v) < budgetEpsilon
}
