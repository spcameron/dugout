package application_test

import (
	"time"

	"github.com/spcameron/dugout/internal/domain"
)

var nyc = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return loc
}()

var todayLock = time.Date(
	1986,
	time.October,
	26,
	0, 0, 0, 0,
	nyc,
)

const (
	teamA = domain.TeamID(999)
	teamB = domain.TeamID(111)
)
