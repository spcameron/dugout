package testkit

import (
	"errors"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
	"github.com/spcameron/dugout/internal/ports"
)

var (
	ErrFailingLoad   = errors.New("load failed")
	ErrFailingAppend = errors.New("append failed")
)

type FailingLoadRosterStore struct{}

func (s *FailingLoadRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	return nil, 0, ErrFailingLoad
}

func (s *FailingLoadRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	panic("stub: FailingLoadRosterStore.Append() always panics")
}

type FailingAppendRosterStore struct {
	Base *FakeRosterStore
}

func (s *FailingAppendRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	if s.Base == nil {
		panic("FailingAppendRosterStore.Base is nil")
	}
	return s.Base.Load(id)
}

func (s *FailingAppendRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	return 0, ErrFailingAppend
}

type VersionConflictRosterStore struct {
	Base *FakeRosterStore
}

func (s *VersionConflictRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	if s.Base == nil {
		panic("FailingAppendRosterStore.Base is nil")
	}
	return s.Base.Load(id)
}

func (s *VersionConflictRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	return 0, ports.ErrVersionConflict
}
