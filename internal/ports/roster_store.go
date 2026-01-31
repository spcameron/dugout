package ports

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/eventlog"
)

type Version eventlog.Sequence

type RosterStore interface {
	Load(id domain.TeamID) ([]eventlog.Recorded[domain.RosterEvent], Version, error)
	Append(id domain.TeamID, newEvents []domain.RosterEvent, expected Version) (Version, error)
}
