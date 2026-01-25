package application

import (
	"errors"

	"github.com/spcameron/dugout/internal/domain"
)

var ErrUnrecognizedRecordedEvent = errors.New("unrecognized recorded event")

type RecordedRosterEvent struct {
	Sequence int
	Event    domain.RosterEvent
}
