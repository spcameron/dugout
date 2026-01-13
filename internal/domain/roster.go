package domain

import "errors"

const (
	MaxRosterSize     = 26
	MaxActiveHitters  = 12
	MaxActivePitchers = 6
)

var (
	ErrActiveHittersFull     = errors.New("roster already has the maximum active hitters")
	ErrActivePitchersFull    = errors.New("roster already has the maximum active pitchers")
	ErrPlayerAlreadyOnRoster = errors.New("player already on roster")
	ErrRosterFull            = errors.New("roster is already full")
	ErrPlayerNotOnRoster     = errors.New("player is not on the roster")
)

type Roster struct {
	TeamID  TeamID
	Entries []RosterEntry
}

func CanAddPlayer(r Roster, id PlayerID) error {
	if len(r.Entries) >= MaxRosterSize {
		return ErrRosterFull
	}

	for _, e := range r.Entries {
		if e.PlayerID == id {
			return ErrPlayerAlreadyOnRoster
		}
	}

	return nil
}

func CanActivatePlayer(r Roster, id PlayerID, role PlayerRole) error {
	var (
		onRoster       bool
		activeHitters  int
		activePitchers int
		inactive       int
	)

	for _, e := range r.Entries {
		if e.PlayerID == id {
			onRoster = true
		}
		switch e.RosterStatus {
		case StatusActiveHitter:
			activeHitters++
		case StatusActivePitcher:
			activePitchers++
		case StatusInactive:
			inactive++
		}
	}

	if !onRoster {
		return ErrPlayerNotOnRoster
	}

	if role == RoleHitter && activeHitters >= MaxActiveHitters {
		return ErrActiveHittersFull
	}

	if role == RolePitcher && activePitchers >= MaxActivePitchers {
		return ErrActivePitchersFull
	}

	return nil
}
