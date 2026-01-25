package testkit

import (
	"time"
)

var nyc = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return loc
}()

// TodayLock returns the date for Game 6, 1986 World Series.
func TodayLock() time.Time {
	return time.Date(
		1986,
		time.October,
		26,
		0, 0, 0, 0,
		nyc,
	)
}

// TomorrowLock returns the date for Game 7, 1986 World Series (Let's Go Mets!).
func TomorrowLock() time.Time {
	return time.Date(
		1986,
		time.October,
		27,
		0, 0, 0, 0,
		nyc,
	)
}
