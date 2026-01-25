package domain

import "time"

type DomainEvent interface {
	isDomainEvent()
}

type RosterEvent interface {
	DomainEvent
	Team() TeamID
	OccurredAt() time.Time
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
func (e AddedPlayerToRoster) OccurredAt() time.Time {
	return e.EffectiveAt
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
func (e ActivatedPlayerOnRoster) OccurredAt() time.Time {
	return e.EffectiveAt
}
