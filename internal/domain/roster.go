package domain

import "errors"

var (
	ErrEventOutsideViewWindow = errors.New("event is outside view effective window")
	ErrWrongTeamID            = errors.New("team IDs do not match")
)

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

	r.EventHistory = append(r.EventHistory, events...)

	return nil
}
