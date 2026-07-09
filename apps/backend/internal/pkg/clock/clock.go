package clock

import "time"

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func System() Clock { return systemClock{} }

func Fixed(t time.Time) Clock { return fixedClock{t: t} }

func OrDefault(clk Clock) Clock {
	if clk == nil {
		return System()
	}
	return clk
}
