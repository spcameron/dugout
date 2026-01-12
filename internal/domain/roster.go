package domain

import "errors"

const MaxRosterSize = 26

var (
	ErrPlayerAlreadyOnRoster = errors.New("player already on roster")
	ErrRosterFull            = errors.New("roster is already full")
)

type Roster struct {
	TeamID  TeamID
	Entries []RosterEntry
}

func CanAddPlayer(r Roster, mlbID MLBPlayerID) error {
	if len(r.Entries) >= MaxRosterSize {
		return ErrRosterFull
	}

	for _, e := range r.Entries {
		if e.MLBID == mlbID {
			return ErrPlayerAlreadyOnRoster
		}
	}

	return nil
}
