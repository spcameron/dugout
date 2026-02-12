package testkit

import (
	"fmt"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
	"github.com/spcameron/dugout/internal/ports"
)

type FakeRosterStore struct {
	committed map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent]
}

func (s *FakeRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	history := s.committed[id]
	history = append([]eventlog.Recorded[domain.RosterEvent](nil), history...)

	var lastSeq eventlog.Sequence
	if len(history) > 0 {
		lastSeq = history[len(history)-1].Sequence
	}

	return history, ports.Version(lastSeq), nil
}

func (s *FakeRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	history := s.committed[id]
	history = append([]eventlog.Recorded[domain.RosterEvent](nil), history...)

	var lastSeq eventlog.Sequence
	if len(history) > 0 {
		lastSeq = history[len(history)-1].Sequence
	}

	current := ports.Version(lastSeq)
	if current != expected {
		return 0, fmt.Errorf("%w: current - %v, expected - %v", ports.ErrVersionConflict, current, expected)
	}

	appendedHistory := make([]eventlog.Recorded[domain.RosterEvent], len(newEvents))
	nextSeq := lastSeq
	for i, ev := range newEvents {
		nextSeq++
		appendedHistory[i] = eventlog.Recorded[domain.RosterEvent]{
			Sequence: nextSeq,
			Event:    ev,
		}
	}

	s.committed[id] = append(history, appendedHistory...)
	newLastSeq := nextSeq

	return ports.Version(newLastSeq), nil
}

// SeedEvents overwrites the entire event stream for the given team,
// then assigns contiguous 1-based sequence numbers to events in the order provided.
func (s *FakeRosterStore) SeedEvents(id domain.TeamID, events []domain.RosterEvent) {
	if s.committed == nil {
		s.committed = make(map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent])
	}

	s.committed[id] = make([]eventlog.Recorded[domain.RosterEvent], len(events))
	for i, ev := range events {
		s.committed[id][i] = eventlog.Recorded[domain.RosterEvent]{
			Sequence: eventlog.Sequence(i + 1),
			Event:    ev,
		}
	}
}

func NewFakeRosterStore() *FakeRosterStore {
	return &FakeRosterStore{
		committed: make(map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent]),
	}
}

func GenerateRosterHistory(id domain.TeamID, players int) []domain.RosterEvent {
	history := make([]domain.RosterEvent, players)
	for i := range players {
		history[i] = domain.AddedPlayerToRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(i + 1),
			EffectiveAt: TodayLock(),
		}
	}

	return history
}

func GenerateActivatedHistory(id domain.TeamID, hitters, pitchers, totalPlayers int) []domain.RosterEvent {
	hitters = min(hitters, domain.MaxActiveHitters)
	pitchers = min(pitchers, domain.MaxActivePitchers)
	totalPlayers = min(totalPlayers, domain.MaxRosterSize)

	totalEvents := hitters + pitchers + totalPlayers
	history := make([]domain.RosterEvent, totalEvents)

	i := 0
	playerID := 1

	for range hitters {
		history[i] = domain.AddedPlayerToRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(playerID),
			EffectiveAt: TodayLock(),
		}

		i++

		history[i] = domain.ActivatedPlayerOnRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(playerID),
			PlayerRole:  domain.RoleHitter,
			EffectiveAt: TodayLock(),
		}

		i++
		playerID++
	}

	for range pitchers {
		history[i] = domain.AddedPlayerToRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(playerID),
			EffectiveAt: TodayLock(),
		}

		i++

		history[i] = domain.ActivatedPlayerOnRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(playerID),
			PlayerRole:  domain.RolePitcher,
			EffectiveAt: TodayLock(),
		}

		i++
		playerID++
	}

	for playerID <= totalPlayers {
		history[i] = domain.AddedPlayerToRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(playerID),
			EffectiveAt: TodayLock(),
		}

		i++
		playerID++
	}

	return history
}
