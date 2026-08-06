package budget

// BudgetExhausted returns true if consumed >= budget.
func BudgetExhausted(consumed, budget float64) bool {
	return consumed >= budget
}
