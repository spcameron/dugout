package application

import (
	"errors"

	"github.com/spcameron/dugout/internal/domain"
)

var ErrUnrecognizedRecordedEvent = errors.New("unrecognized recorded event")

type RecordedEvent struct {
	Sequence int
	Event    domain.DomainEvent
}
