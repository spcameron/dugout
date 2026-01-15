package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	MaxRosterSize     = 26
	MaxActiveHitters  = 12
	MaxActivePitchers = 6
)

var (
	ErrActiveHittersFull     = errors.New("roster already has the maximum active hitters")
	ErrActivePitchersFull    = errors.New("roster already has the maximum active pitchers")
	ErrPlayerAlreadyActive   = errors.New("player already activated")
	ErrPlayerAlreadyOnRoster = errors.New("player already on roster")
	ErrRosterFull            = errors.New("roster is already full")
	ErrPlayerNotOnRoster     = errors.New("player is not on the roster")
)

type RosterCounts struct {
	Total          int
	ActiveHitters  int
	ActivePitchers int
	Inactive       int
}

type RosterView struct {
	TeamID           TeamID
	Entries          []RosterEntry
	EffectiveThrough time.Time
}

func (r RosterView) Counts() RosterCounts {
	rc := RosterCounts{}

	for _, e := range r.Entries {
		switch e.RosterStatus {
		case StatusActiveHitter:
			rc.ActiveHitters++
		case StatusActivePitcher:
			rc.ActivePitchers++
		case StatusInactive:
			rc.Inactive++
		default:
			panic(fmt.Errorf("unrecognized roster status: %v", e.RosterStatus))
		}

		rc.Total++
	}

	return rc
}

func (r RosterView) DecideAddPlayer(id PlayerID, effectiveAt time.Time) ([]DomainEvent, error) {
	err := r.CanAddPlayer(id)
	if err != nil {
		return nil, err
	}

	res := []DomainEvent{
		AddedPlayerToRoster{
			PlayerID:    id,
			EffectiveAt: effectiveAt,
		},
	}

	return res, nil
}

func (r RosterView) CanAddPlayer(id PlayerID) error {
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

func (r RosterView) CanActivatePlayer(id PlayerID, role PlayerRole) error {
	var onRoster bool

	for _, e := range r.Entries {
		if e.PlayerID == id {
			if e.RosterStatus != StatusInactive {
				return ErrPlayerAlreadyActive
			}

			onRoster = true
			break
		}
	}

	if !onRoster {
		return ErrPlayerNotOnRoster
	}

	rc := r.Counts()

	if role == RoleHitter && rc.ActiveHitters >= MaxActiveHitters {
		return ErrActiveHittersFull
	}
	if role == RolePitcher && rc.ActivePitchers >= MaxActivePitchers {
		return ErrActivePitchersFull
	}

	return nil
}

type AddedPlayerToRoster struct {
	PlayerID    PlayerID
	EffectiveAt time.Time
}

func (e AddedPlayerToRoster) isDomainEvent() {}
