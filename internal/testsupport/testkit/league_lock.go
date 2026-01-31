package testkit

import "time"

type StubLeagueLock struct {
	Last time.Time
	Next time.Time
}

func (s StubLeagueLock) LastLock() time.Time {
	return s.Last
}

func (s StubLeagueLock) NextLock() time.Time {
	return s.Next
}

func NewStubLeagueLock() StubLeagueLock {
	return StubLeagueLock{
		Last: TodayLock(),
		Next: TomorrowLock(),
	}
}
