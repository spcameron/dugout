package domain

type TeamID int

type DomainEvent interface {
	isDomainEvent()
}
