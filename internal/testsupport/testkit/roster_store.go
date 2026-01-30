package testkit

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
	"github.com/spcameron/dugout/internal/ports"
)

type FakeRosterStore struct {
	committed map[domain.TeamID][]eventlog.Recorded[domain.RosterEvent]
}

func (s FakeRosterStore) Load(id domain.TeamID) (committed []eventlog.Recorded[domain.RosterEvent], version ports.Version, err error) {
	return nil, 0, nil
}

func (s FakeRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (newVersion ports.Version, err error) {
	return 0, nil
}

func NewFakeRosterStore() *FakeRosterStore {
	s := FakeRosterStore{}

	return &s
}
