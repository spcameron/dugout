package domain

type TeamID int

// TODO: extract to events.go file
type DomainEvent interface {
	isDomainEvent()
}
