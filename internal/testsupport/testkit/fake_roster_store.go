package testkit

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
	"github.com/spcameron/dugout/internal/ports"
)

type FakeRosterStore struct {
	committed map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent]
}

func (s *FakeRosterStore) Load(id domain.TeamID) (committed []eventlog.Recorded[domain.RosterEvent], version ports.Version, err error) {
	committed, ok := s.committed[id]
	if !ok {
		return nil, 0, nil
	}

	for _, r := range committed {
		version = max(version, ports.Version(r.Sequence))
	}

	return committed, version, nil
}

func (s *FakeRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (newVersion ports.Version, err error) {
	return 0, nil
}

func NewFakeRosterStore() *FakeRosterStore {
	s := FakeRosterStore{
		committed: make(map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent]),
	}

	s.committed[TeamA()] = []eventlog.Recorded[domain.RosterEvent]{}

	return &s
}
