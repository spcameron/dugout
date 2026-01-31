package eventlog

import "errors"

var (
	ErrUnrecognizedRecordedEvent      = errors.New("unrecognized recorded event")
	ErrDuplicateRecordedEventSequence = errors.New("duplicate recorded event sequence")
)

type Sequence int64

type Recorded[E any] struct {
	Sequence Sequence
	Event    E
}
