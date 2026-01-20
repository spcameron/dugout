package domain

import "errors"

var ErrWrongTeamID = errors.New("team IDs do not match")

type Roster struct {
	TeamID       TeamID
	EventHistory []RosterEvent
}

func (r *Roster) Append(events ...RosterEvent) error {
	for _, ev := range events {
		if ev.Team() != r.TeamID {
			return ErrWrongTeamID
		}
	}

	for _, ev := range events {
		r.EventHistory = append(r.EventHistory, ev)
	}

	// later: r.EventHistory = append(r.EventHistory, events...)

	return nil
}
