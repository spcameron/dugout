package testkit

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
	"github.com/spcameron/dugout/internal/ports"
)

// SpyRosterStore must be constructed with NewSpyRosterStore.
type SpyRosterStore struct {
	InnerStore  ports.RosterStore
	LoadCalls   []domain.TeamID
	AppendCalls []struct {
		TeamID  domain.TeamID
		Events  []domain.RosterEvent
		Version ports.Version
	}
}

func (s *SpyRosterStore) Load(id domain.TeamID) (committed []eventlog.Recorded[domain.RosterEvent], version ports.Version, err error) {
	s.LoadCalls = append(s.LoadCalls, id)
	return s.InnerStore.Load(id)
}

func (s *SpyRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (newVersion ports.Version, err error) {
	eventsCopy := append([]domain.RosterEvent(nil), newEvents...)
	s.AppendCalls = append(s.AppendCalls, struct {
		TeamID  domain.TeamID
		Events  []domain.RosterEvent
		Version ports.Version
	}{
		TeamID:  id,
		Events:  eventsCopy,
		Version: expected,
	})

	return s.InnerStore.Append(id, eventsCopy, expected)
}

func NewSpyRosterStore(inner ports.RosterStore) *SpyRosterStore {
	if inner == nil {
		panic("SpyRosterStore.InnerStore is nil")
	}

	return &SpyRosterStore{
		InnerStore: inner,
	}
}
