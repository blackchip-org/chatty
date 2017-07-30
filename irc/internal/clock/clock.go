package clock

import "time"

type C interface {
	Now() time.Time
}

type Real struct{}

func (r Real) Now() time.Time {
	return time.Now()
}

type Mock struct {
	now time.Time
}

func (m Mock) Now() time.Time {
	return m.now
}

func (m *Mock) Add(d time.Duration) {
	m.now = m.now.Add(d)
}
