package testkit

import "github.com/spcameron/dugout/internal/domain"

// TeamA returns the fixed TeamID 999.
func TeamA() domain.TeamID {
	return domain.TeamID(999)
}

// TeamB returns the fixed TeamID 111.
func TeamB() domain.TeamID {
	return domain.TeamID(111)
}
