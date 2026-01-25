package domain

import (
	"time"
)

type Roster struct {
	TeamID       TeamID
	EventHistory []RosterEvent
}

// func (r *Roster) Append(events ...RosterEvent) error {
// 	for _, ev := range events {
// 		if ev.Team() != r.TeamID {
// 			return ErrWrongTeamID
// 		}
// 	}
//
// 	r.EventHistory = append(r.EventHistory, events...)
//
// 	return nil
// }

func (r Roster) ProjectThrough(through time.Time) RosterView {
	rv := RosterView{
		TeamID:           r.TeamID,
		EffectiveThrough: through,
	}

	for _, e := range r.EventHistory {
		if e.OccurredAt().After(through) {
			continue
		}

		rv.Apply(e)
	}

	return rv
}
