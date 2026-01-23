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
	ErrActiveHittersFull        = errors.New("roster already has the maximum active hitters")
	ErrActivePitchersFull       = errors.New("roster already has the maximum active pitchers")
	ErrPlayerAlreadyActive      = errors.New("player already activated")
	ErrPlayerAlreadyOnRoster    = errors.New("player already on roster")
	ErrRosterFull               = errors.New("roster is already full")
	ErrPlayerNotOnRoster        = errors.New("player is not on the roster")
	ErrUnrecognizedPlayerRole   = errors.New("unrecognized player role")
	ErrUnrecognizedRosterEvent  = errors.New("unrecognized roster event")
	ErrUnrecognizedRosterStatus = errors.New("unrecognized roster status")
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

func (rv RosterView) Counts() RosterCounts {
	rc := RosterCounts{}

	for _, e := range rv.Entries {
		switch e.RosterStatus {
		case StatusActiveHitter:
			rc.ActiveHitters++
		case StatusActivePitcher:
			rc.ActivePitchers++
		case StatusInactive:
			rc.Inactive++
		default:
			panic(fmt.Errorf("%w: %v", ErrUnrecognizedRosterStatus, e.RosterStatus))
		}

		rc.Total++
	}

	return rc
}

// DecideAddPlayer returns the AddedPlayerToRoster events that should be recorded if allowed.
func (rv RosterView) DecideAddPlayer(id PlayerID, effectiveAt time.Time) ([]RosterEvent, error) {
	err := rv.validateAddPlayer(id)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		AddedPlayerToRoster{
			PlayerID:    id,
			EffectiveAt: effectiveAt,
		},
	}

	return res, nil
}

// DecideActivatePlayer returns the ActivatedPlayerOnRoster events that should be recorded if allowed.
func (rv RosterView) DecideActivatePlayer(id PlayerID, role PlayerRole, effectiveAt time.Time) ([]RosterEvent, error) {
	err := rv.validateActivatePlayer(id, role)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		ActivatedPlayerOnRoster{
			PlayerID:    id,
			PlayerRole:  role,
			EffectiveAt: effectiveAt,
		},
	}

	return res, nil
}

func (rv *RosterView) Apply(event RosterEvent) {
	if event.OccurredAt().After(rv.EffectiveThrough) {
		panic(fmt.Errorf("%w: event lock %v, view lock %v", ErrEventOutsideViewWindow, event.OccurredAt(), rv.EffectiveThrough))
	}

	if rv.TeamID != event.Team() {
		panic(fmt.Errorf("%w: event team %v, view team %v", ErrWrongTeamID, event.Team(), rv.TeamID))
	}

	switch ev := event.(type) {
	case AddedPlayerToRoster:
		if rv.PlayerOnRoster(ev.PlayerID) {
			panic(fmt.Errorf("%w: player ID %v", ErrPlayerAlreadyOnRoster, ev.PlayerID))
		}

		rv.Entries = append(rv.Entries, RosterEntry{
			PlayerID:     ev.PlayerID,
			RosterStatus: StatusInactive,
		})
	default:
		panic(fmt.Errorf("%w: %T", ErrUnrecognizedRosterEvent, event))
	}
}

func (rv RosterView) PlayerOnRoster(id PlayerID) bool {
	for _, e := range rv.Entries {
		if e.PlayerID == id {
			return true
		}
	}

	return false
}

func (rv RosterView) validateAddPlayer(id PlayerID) error {
	if len(rv.Entries) >= MaxRosterSize {
		return ErrRosterFull
	}

	if rv.PlayerOnRoster(id) {
		return ErrPlayerAlreadyOnRoster
	}

	return nil
}

func (rv RosterView) validateActivatePlayer(id PlayerID, role PlayerRole) error {
	var onRoster bool
	for _, e := range rv.Entries {
		if e.PlayerID == id {
			if e.RosterStatus == StatusActiveHitter || e.RosterStatus == StatusActivePitcher {
				return ErrPlayerAlreadyActive
			}

			onRoster = true
			break
		}
	}

	if !onRoster {
		return ErrPlayerNotOnRoster
	}

	rc := rv.Counts()

	switch role {
	case RoleHitter:
		if rc.ActiveHitters >= MaxActiveHitters {
			return ErrActiveHittersFull
		}
	case RolePitcher:
		if rc.ActivePitchers >= MaxActivePitchers {
			return ErrActivePitchersFull
		}
	default:
		return ErrUnrecognizedPlayerRole
	}

	return nil
}
