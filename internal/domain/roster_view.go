package domain

import (
	"fmt"
	"time"
)

const (
	MaxRosterSize     = 26
	MaxActiveHitters  = 12
	MaxActivePitchers = 6
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

// Counts tabulates the number of total players, active hitters, active pitchers,
// and inactive players among the stored RosterEntries.
//
// Panics if an unrecognized RosterStatus is encountered.
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
func (rv RosterView) DecideAddPlayer(id PlayerID) ([]RosterEvent, error) {
	err := rv.validateAddPlayer(id)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		AddedPlayerToRoster{
			TeamID:      rv.TeamID,
			PlayerID:    id,
			EffectiveAt: rv.EffectiveThrough,
		},
	}

	return res, nil
}

// DecideRemovePlayer returns the RemovedPlayerFromRoster events that should be recorded if allowed.
func (rv RosterView) DecideRemovePlayer(id PlayerID) ([]RosterEvent, error) {
	err := rv.validateRemovePlayer(id)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		RemovedPlayerFromRoster{
			TeamID:      rv.TeamID,
			PlayerID:    id,
			EffectiveAt: rv.EffectiveThrough,
		},
	}

	return res, nil
}

// DecideActivatePlayer returns the ActivatedPlayerOnRoster events that should be recorded if allowed.
func (rv RosterView) DecideActivatePlayer(id PlayerID, role PlayerRole) ([]RosterEvent, error) {
	err := rv.validateActivatePlayer(id, role)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		ActivatedPlayerOnRoster{
			TeamID:      rv.TeamID,
			PlayerID:    id,
			PlayerRole:  role,
			EffectiveAt: rv.EffectiveThrough,
		},
	}

	return res, nil
}

// DecideInactivatePlayer returns the InactivatedPlayerOnRoster events that should be recorded if allowed.
func (rv RosterView) DecideInactivatePlayer(id PlayerID) ([]RosterEvent, error) {
	err := rv.validateInactivatePlayer(id)
	if err != nil {
		return nil, err
	}

	res := []RosterEvent{
		InactivatedPlayerOnRoster{
			TeamID:      rv.TeamID,
			PlayerID:    id,
			EffectiveAt: rv.EffectiveThrough,
		},
	}

	return res, nil
}

// Apply applies a roster domain event to the view.
//
// Events whose postconditions already hold are treated as no-ops.
// Events that would violate roster invariants cause Apply to panic.
func (rv *RosterView) Apply(event RosterEvent) {
	if event.OccurredAt().After(rv.EffectiveThrough) {
		panic(fmt.Errorf("%w: event lock %v, view lock %v", ErrEventOutsideViewWindow, event.OccurredAt(), rv.EffectiveThrough))
	}

	if rv.TeamID != event.Team() {
		panic(fmt.Errorf("%w: event team %v, view team %v", ErrWrongTeamID, event.Team(), rv.TeamID))
	}

	switch ev := event.(type) {
	case AddedPlayerToRoster:
		rv.addPlayer(ev.PlayerID)
	case RemovedPlayerFromRoster:
		rv.removePlayer(ev.PlayerID)
	case ActivatedPlayerOnRoster:
		rv.activatePlayer(ev.PlayerID, ev.PlayerRole)
	case InactivatedPlayerOnRoster:
		rv.inactivatePlayer(ev.PlayerID)
	default:
		panic(fmt.Errorf("%w: %T", ErrUnrecognizedRosterEvent, event))
	}
}

// PlayerOnRoster checks the PlayerID for each RosterEntry, and returns true if
// a match is found for the PlayerID passed as argument.
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

func (rv RosterView) validateRemovePlayer(id PlayerID) error {
	if !rv.PlayerOnRoster(id) {
		return ErrPlayerNotOnRoster
	}

	return nil
}

func (rv RosterView) validateActivatePlayer(id PlayerID, role PlayerRole) error {
	var onRoster bool
	for _, e := range rv.Entries {
		if e.PlayerID == id {
			if e.RosterStatus == StatusInactive {
				onRoster = true
				break
			}

			if e.RosterStatus == StatusActiveHitter || e.RosterStatus == StatusActivePitcher {
				return ErrPlayerAlreadyActive
			}

			return ErrUnrecognizedRosterStatus
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

func (rv RosterView) validateInactivatePlayer(id PlayerID) error {
	var onRoster bool
	for _, e := range rv.Entries {
		if e.PlayerID == id {
			switch e.RosterStatus {
			case StatusInactive:
				return ErrPlayerAlreadyInactive
			case StatusActiveHitter:
			case StatusActivePitcher:
			default:
				return ErrUnrecognizedRosterStatus
			}

			onRoster = true
			break
		}
	}

	if !onRoster {
		return ErrPlayerNotOnRoster
	}

	return nil
}

func (rv *RosterView) addPlayer(id PlayerID) {
	if rv.PlayerOnRoster(id) {
		panic(fmt.Errorf("%w: player ID %v", ErrPlayerAlreadyOnRoster, id))
	}

	rv.Entries = append(rv.Entries, RosterEntry{
		TeamID:       rv.TeamID,
		PlayerID:     id,
		RosterStatus: StatusInactive,
	})
}

func (rv *RosterView) removePlayer(id PlayerID) {
	for i, e := range rv.Entries {
		if e.PlayerID != id {
			continue
		}

		copy(rv.Entries[i:], rv.Entries[i+1:])

		var zero RosterEntry
		rv.Entries[len(rv.Entries)-1] = zero

		rv.Entries = rv.Entries[:len(rv.Entries)-1]
		return
	}
}

func (rv *RosterView) activatePlayer(id PlayerID, role PlayerRole) {
	for i, e := range rv.Entries {
		if e.PlayerID != id {
			continue
		}

		switch role {
		case RoleHitter:
			rv.Entries[i].RosterStatus = StatusActiveHitter
		case RolePitcher:
			rv.Entries[i].RosterStatus = StatusActivePitcher
		default:
			panic(fmt.Errorf("%w: %s", ErrUnrecognizedPlayerRole, role))
		}
		return
	}

	panic(fmt.Errorf("%w: player ID %v", ErrPlayerNotOnRoster, id))
}

func (rv *RosterView) inactivatePlayer(id PlayerID) {
	for i, e := range rv.Entries {
		if e.PlayerID != id {
			continue
		}

		switch e.RosterStatus {
		case StatusActiveHitter:
			rv.Entries[i].RosterStatus = StatusInactive
		case StatusActivePitcher:
			rv.Entries[i].RosterStatus = StatusInactive
		case StatusInactive:
			return
		default:
			panic(fmt.Errorf("%w: %v", ErrUnrecognizedRosterStatus, e.RosterStatus))
		}

		return
	}

	panic(fmt.Errorf("%w: playerID %v", ErrPlayerNotOnRoster, id))
}
