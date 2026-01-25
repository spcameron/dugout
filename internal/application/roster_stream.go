package application

import (
	"fmt"
	"slices"
	"time"

	"github.com/spcameron/dugout/internal/domain"
)

type RosterStream struct {
	TeamID         domain.TeamID
	RecordedEvents []RecordedRosterEvent
}

func (rs *RosterStream) Append(events ...RecordedRosterEvent) error {
	for _, re := range events {
		if re.Event.Team() != rs.TeamID {
			return fmt.Errorf("%w: event team %v, roster team %v", domain.ErrWrongTeamID, re.Event.Team(), rs.TeamID)
		}
	}

	rs.RecordedEvents = append(rs.RecordedEvents, events...)

	return nil
}

func (rs RosterStream) ProjectThrough(through time.Time) domain.RosterView {
	rv := domain.RosterView{
		TeamID:           rs.TeamID,
		EffectiveThrough: through,
	}

	history := slices.Clone(rs.RecordedEvents)
	for _, ev := range history {
		if ev.Event.OccurredAt().After(through) {
			continue
		}

		rv.Apply(ev.Event)
	}

	return rv
}
