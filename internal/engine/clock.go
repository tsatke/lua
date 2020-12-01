package engine

import "time"

type Clock interface {
	Now() time.Time
}

type sysClock struct{}

func (sysClock) Now() time.Time { return time.Now() }
