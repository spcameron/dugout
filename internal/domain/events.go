package domain

import "time"

type RosterEvent interface {
	isDomainEvent()
	Team() TeamID
}

type AddedPlayerToRoster struct {
	TeamID      TeamID
	PlayerID    PlayerID
	EffectiveAt time.Time
}

func (e AddedPlayerToRoster) isDomainEvent() {}
func (e AddedPlayerToRoster) Team() TeamID {
	return e.TeamID
}

type ActivatedPlayerOnRoster struct {
	TeamID      TeamID
	PlayerID    PlayerID
	PlayerRole  PlayerRole
	EffectiveAt time.Time
}

func (e ActivatedPlayerOnRoster) isDomainEvent() {}
func (e ActivatedPlayerOnRoster) Team() TeamID {
	return e.TeamID
}
