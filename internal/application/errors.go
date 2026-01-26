package application

import "errors"

var (
	ErrUnrecognizedRecordedEvent      = errors.New("unrecognized recorded event")
	ErrDuplicateRecordedEventSequence = errors.New("duplicate recorded event sequence")
)
