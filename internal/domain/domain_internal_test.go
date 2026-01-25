package domain

import (
	"testing"
	"time"

	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
)

type unknownRosterEvent struct {
	TeamID      TeamID
	EffectiveAt time.Time
}

func (e unknownRosterEvent) isDomainEvent() {}
func (e unknownRosterEvent) Team() TeamID {
	return e.TeamID
}
func (e unknownRosterEvent) OccurredAt() time.Time {
	return e.EffectiveAt
}

func teamA() TeamID {
	return TeamID(999)
}

func todayLock() time.Time {
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	return time.Date(
		1986,
		time.October,
		26,
		0, 0, 0, 0,
		nyc,
	)
}

func TestRosterViewApply(t *testing.T) {
	t.Run("apply panics if roster event is unrecognized", func(t *testing.T) {
		rv := RosterView{
			TeamID:           teamA(),
			Entries:          nil,
			EffectiveThrough: todayLock(),
		}

		startingTeamID := rv.TeamID
		startingLength := len(rv.Entries)
		startingLock := rv.EffectiveThrough

		e := unknownRosterEvent{
			TeamID:      teamA(),
			EffectiveAt: todayLock(),
		}

		fn := func() { rv.Apply(e) }

		err := require.PanicsError(t, fn)
		require.NotNil(t, err)
		require.ErrorIs(t, err, ErrUnrecognizedRosterEvent)

		assert.Equal(t, rv.TeamID, startingTeamID)
		assert.Equal(t, len(rv.Entries), startingLength)
		assert.Equal(t, rv.EffectiveThrough, startingLock)
	})
}
