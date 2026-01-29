package roster

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/spcameron/dugout/internal/domain"
)

type RosterStream struct {
	TeamID    domain.TeamID
	Committed []RecordedRosterEvent
	Pending   []domain.RosterEvent
}

func (rs *RosterStream) Stage(events ...domain.RosterEvent) error {
	for _, re := range events {
		if re.Team() != rs.TeamID {
			return fmt.Errorf("%w: event team %v, roster team %v", domain.ErrWrongTeamID, re.Team(), rs.TeamID)
		}
	}

	rs.Pending = append(rs.Pending, events...)

	return nil
}

func (rs RosterStream) ProjectThrough(through time.Time) domain.RosterView {
	rv := domain.RosterView{
		TeamID:           rs.TeamID,
		EffectiveThrough: through,
	}

	sortedCommitted := orderEventsByUniqueSequence(rs.Committed)

	applyThrough(&rv, through, extractEvents(sortedCommitted))
	applyThrough(&rv, through, rs.Pending)

	return rv
}

func orderEventsByUniqueSequence(recordedEvents []RecordedRosterEvent) []RecordedRosterEvent {
	assertUniqueSequences(recordedEvents)

	return sortBySequence(recordedEvents)
}

func assertUniqueSequences(recordedEvents []RecordedRosterEvent) {
	seen := make(map[int64]struct{}, len(recordedEvents))
	for _, ev := range recordedEvents {
		if _, ok := seen[ev.Sequence]; !ok {
			seen[ev.Sequence] = struct{}{}
			continue
		}

		panic(fmt.Errorf("%w: %v", ErrDuplicateRecordedEventSequence, ev.Sequence))
	}
}

func sortBySequence(recordedEvents []RecordedRosterEvent) []RecordedRosterEvent {
	sorted := make([]RecordedRosterEvent, len(recordedEvents))
	copy(sorted, recordedEvents)
	slices.SortFunc(sorted, func(a, b RecordedRosterEvent) int {
		return cmp.Compare(a.Sequence, b.Sequence)
	})

	return sorted
}

func extractEvents(recordedEvents []RecordedRosterEvent) []domain.RosterEvent {
	extracted := make([]domain.RosterEvent, len(recordedEvents))
	for i, re := range recordedEvents {
		extracted[i] = re.Event
	}

	return extracted
}

func applyThrough(rv *domain.RosterView, through time.Time, events []domain.RosterEvent) {
	for _, ev := range events {
		if ev.OccurredAt().After(through) {
			continue
		}

		rv.Apply(ev)
	}
}
