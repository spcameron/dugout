package domain

type PlayerRole int

const (
	RoleHitter PlayerRole = iota + 1
	RolePitcher
)

type PlayerID int
type MLBPlayerID int

type Player struct {
	ID    PlayerID
	MLBID MLBPlayerID
	Name  string
	Role  PlayerRole
}
