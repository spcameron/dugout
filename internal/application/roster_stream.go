package application

import (
	"fmt"

	"github.com/spcameron/dugout/internal/domain"
)

type RosterStream struct {
	TeamID       domain.TeamID
	EventHistory []RecordedEvent
}

func (rs *RosterStream) Append(events ...RecordedEvent) error {
	for _, re := range events {
		ev, ok := re.Event.(domain.RosterEvent)
		if !ok {
			return fmt.Errorf("%w: %T", ErrUnrecognizedRecordedEvent, re.Event)
		}

		if ev.Team() != rs.TeamID {
			return fmt.Errorf("%w: event team %v, roster team %v", domain.ErrWrongTeamID, ev.Team(), rs.TeamID)
		}
	}

	rs.EventHistory = append(rs.EventHistory, events...)

	return nil
}
