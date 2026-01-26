package application

import (
	"cmp"
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

	includedEvents := rs.filterAndSortHistory(through)
	for _, ev := range includedEvents {
		rv.Apply(ev.Event)
	}

	return rv
}

func (rs RosterStream) filterAndSortHistory(through time.Time) []RecordedRosterEvent {
	res := []RecordedRosterEvent{}

	for _, ev := range rs.RecordedEvents {
		if ev.Event.OccurredAt().After(through) {
			continue
		}

		res = append(res, ev)
	}

	seen := make(map[int64]struct{}, len(res))
	for _, ev := range res {
		if _, ok := seen[ev.Sequence]; !ok {
			seen[ev.Sequence] = struct{}{}
			continue
		}

		panic(fmt.Errorf("%w: %v", ErrDuplicateRecordedEventSequence, ev.Sequence))
	}

	slices.SortFunc(res, func(a, b RecordedRosterEvent) int {
		return cmp.Compare(a.Sequence, b.Sequence)
	})

	return res
}
