//go:build testhook

package budget

// SetLastMonthForTest overrides month tracking for tests (requires -tags=testhook).
func SetLastMonthForTest(s *MonthlyRebalanceScheduler, month string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastMonth = month
}
