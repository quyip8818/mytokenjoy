package points

import "github.com/tokenjoy/backend/internal/pkg/exchange"

func FromDisplay(display float64) float64 {
	return exchange.ToPoints(display)
}

func ToDisplay(points float64) float64 {
	return exchange.ToDisplay(points)
}
