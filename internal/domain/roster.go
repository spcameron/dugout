package domain

import "errors"

var ErrPlayerAlreadyOnRoster = errors.New("player already on roster")

type Roster struct {
	TeamID  TeamID
	Entries []RosterEntry
}

func CanAddPlayer(r Roster, mlbID MLBPlayerID) error {
	for _, e := range r.Entries {
		if e.MLBID == mlbID {
			return ErrPlayerAlreadyOnRoster
		}
	}

	return nil
}
