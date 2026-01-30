package ports

import "time"

type LeagueLock interface {
	LastLock() time.Time
	NextLock() time.Time
}
