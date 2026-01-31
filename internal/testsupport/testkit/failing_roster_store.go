package testkit

import (
	"errors"
	"fmt"

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

type FailingAppendRosterStore struct{}

func (s *FailingAppendRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	return nil, 0, nil
}

func (s *FailingAppendRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	return 0, ErrFailingAppend
}

type VersionConflictRosterStore struct{}

func (s *VersionConflictRosterStore) Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], ports.Version, error) {
	return nil, 1, nil
}

func (s *VersionConflictRosterStore) Append(id domain.TeamID, newEvents []domain.RosterEvent, expected ports.Version) (ports.Version, error) {
	current := expected + 1
	if current != expected {
		return 0, fmt.Errorf("%w: current - %v, expected - %v", ports.ErrVersionConflict, current, expected)
	}

	return 0, nil
}
