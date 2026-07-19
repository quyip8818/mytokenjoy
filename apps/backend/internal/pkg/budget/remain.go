package budget

// BudgetExhausted returns true if consumed >= budget.
func BudgetExhausted(consumed, budget int64) bool {
	return consumed >= budget
}
