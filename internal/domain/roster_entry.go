package domain

type RosterStatus int

const (
	StatusInactive RosterStatus = iota + 1
	StatusActiveHitter
	StatusActivePitcher
)

type RosterEntry struct {
	PlayerID     PlayerID
	RosterStatus RosterStatus
}
