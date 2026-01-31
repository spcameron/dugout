package testkit

import "github.com/spcameron/dugout/internal/domain"

// TeamA returns the fixed TeamID 111.
func TeamA() domain.TeamID {
	return domain.TeamID(111)
}

// TeamB returns the fixed TeamID 222.
func TeamB() domain.TeamID {
	return domain.TeamID(222)
}

// TeamC returns the fixed TeamID 333.
func TeamC() domain.TeamID {
	return domain.TeamID(333)
}
