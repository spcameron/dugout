package domain

import "fmt"

type RosterStatus int

const (
	StatusInactive RosterStatus = iota + 1
	StatusActiveHitter
	StatusActivePitcher
)

func (s RosterStatus) String() string {
	switch s {
	case StatusInactive:
		return "StatusInactive"
	case StatusActiveHitter:
		return "StatusActiveHitter"
	case StatusActivePitcher:
		return "StatusActivePitcher"
	default:
		return fmt.Sprintf("RosterStatus(%d)", int(s))
	}
}

type RosterEntry struct {
	TeamID       TeamID
	PlayerID     PlayerID
	RosterStatus RosterStatus
}
