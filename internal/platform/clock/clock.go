package clock

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

type MockClock struct {
	now time.Time
}

func NewMockClock(now time.Time) *MockClock {
	return &MockClock{now: now.UTC()}
}

func (m *MockClock) Now() time.Time {
	return m.now
}

func (m *MockClock) Add(d time.Duration) {
	m.now = m.now.Add(d)
}

func (m *MockClock) Set(t time.Time) {
	m.now = t.UTC()
}
