package domain

import "time"

type DomainEvent interface {
	isDomainEvent()
}

type AddedPlayerToRoster struct {
	PlayerID    PlayerID
	EffectiveAt time.Time
}

func (e AddedPlayerToRoster) isDomainEvent() {}

type ActivatedPlayerOnRoster struct {
	PlayerID     PlayerID
	RosterStatus RosterStatus
	EffectiveAt  time.Time
}

func (e ActivatedPlayerOnRoster) isDomainEvent() {}
