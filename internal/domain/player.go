package domain

type PlayerID int
type MLBPlayerID int

type Player struct {
	ID    PlayerID
	MLBID MLBPlayerID
	Name  string
}
