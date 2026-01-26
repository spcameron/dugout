package application

import (
	"github.com/spcameron/dugout/internal/domain"
)

type RecordedRosterEvent struct {
	Sequence int64
	Event    domain.RosterEvent
}
