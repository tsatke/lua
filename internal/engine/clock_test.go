package engine

import "time"

type mockClock struct{}

func (m mockClock) Now() time.Time {
	return time.Unix(1606850863, 419123456) // 2020-12-01 19:27:43.419123456 +0000 UTC
}
