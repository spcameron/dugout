package roster

import (
	"errors"

	"github.com/spcameron/dugout/internal/domain"
)

var (
	ErrUnrecognizedRecordedEvent      = errors.New("unrecognized recorded event")
	ErrDuplicateRecordedEventSequence = errors.New("duplicate recorded event sequence")
)

type RecordedRosterEvent struct {
	Sequence int64
	Event    domain.RosterEvent
}
