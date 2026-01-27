package testkit

import (
	"slices"
	"time"

	"github.com/spcameron/dugout/internal/domain"
)

// NewRosterView returns a RosterView containing a given number of players.
//
// Players will be assigned consecutive PlayerIDs beginning from 1, and an inactive RosterStatus.
// Panics if the number of players is less than zero, or greater than MaxRosterSize.
func NewRosterView(teamID domain.TeamID, players int, lock time.Time) domain.RosterView {
	if players < 0 {
		panic("number of players cannot be negative")
	}

	if players > domain.MaxRosterSize {
		panic("number of players cannot exceed MaxRosterSize")
	}

	rv := domain.RosterView{
		TeamID:           teamID,
		Entries:          make([]domain.RosterEntry, players),
		EffectiveThrough: lock,
	}

	for i := range players {
		rv.Entries[i] = domain.RosterEntry{
			TeamID:       teamID,
			PlayerID:     domain.PlayerID(i + 1),
			RosterStatus: domain.StatusInactive,
		}
	}

	return rv
}

// ActivatedRosterView returns a RosterView with a given number of active hitters and active pitchers.
//
// The number of hitters and pitchers will never exceed the MaxActiveHitters and MaxActivePitchers.
// Creates a shallow copy of the RosterView.Entries slice, so shared ownership is safe.
// Panics if the given number of hitters and pitchers exceeds the length of rv.Entries.
func ActivatedRosterView(rv domain.RosterView, hitters, pitchers int) domain.RosterView {
	if len(rv.Entries) < hitters+pitchers {
		panic("roster entries cannot be fewer than total hitters and pitchers")
	}

	hitters = min(hitters, domain.MaxActiveHitters)
	pitchers = min(pitchers, domain.MaxActivePitchers)

	rv.Entries = slices.Clone(rv.Entries)

	i := 0
	for range hitters {
		rv.Entries[i].RosterStatus = domain.StatusActiveHitter
		i++
	}
	for range pitchers {
		rv.Entries[i].RosterStatus = domain.StatusActivePitcher
		i++
	}

	return rv
}
