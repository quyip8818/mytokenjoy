package newapiunits

import "math"

// QuotaDelta returns target-current without wrapping, clamped so current+delta stays in [0, MaxInt64].
func QuotaDelta(target, current int64) int64 {
	if current < 0 {
		current = 0
	}
	if target < 0 {
		target = 0
	}
	if target == current {
		return 0
	}
	if target > current {
		need := target - current
		room := int64(math.MaxInt64) - current
		if need > room {
			return room
		}
		return need
	}
	return target - current
}

// AddSat returns a+b saturating at MaxInt64 / MinInt64 (no wrap).
func AddSat(a, b int64) int64 {
	if b > 0 {
		if a > math.MaxInt64-b {
			return math.MaxInt64
		}
		return a + b
	}
	if b < 0 {
		if a < math.MinInt64-b {
			return math.MinInt64
		}
		return a + b
	}
	return a
}

// SubFloor0 returns max(a-b, 0) without underflow wrap.
func SubFloor0(a, b int64) int64 {
	if a <= 0 {
		return 0
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return 0
	}
	return a - b
}
