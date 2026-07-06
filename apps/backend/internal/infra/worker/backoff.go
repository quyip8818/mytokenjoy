package worker

import (
	"math"
	"time"
)

func backoff(attempts int) time.Duration {
	seconds := math.Min(300, math.Pow(2, float64(attempts)))
	return time.Duration(seconds) * time.Second
}
