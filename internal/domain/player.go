package domain

import "fmt"

type PlayerRole int

const (
	RoleHitter PlayerRole = iota + 1
	RolePitcher
)

func (s PlayerRole) String() string {
	switch s {
	case RoleHitter:
		return "RoleHitter"
	case RolePitcher:
		return "RolePitcher"
	default:
		return fmt.Sprintf("PlayerRole(%d)", int(s))
	}
}

type PlayerID int
type MLBPlayerID int

type Player struct {
	ID    PlayerID
	MLBID MLBPlayerID
	Name  string
	Role  PlayerRole
}
