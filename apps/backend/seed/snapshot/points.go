package snapshot

import "github.com/tokenjoy/backend/seed/points"

func seedPoints(display float64) float64 {
	return points.FromDisplay(display)
}
