package domain_test

import (
	"errors"
	"testing"

	"github.com/spcameron/dugout/internal/domain"
)

func TestCanAddPlayer(t *testing.T) {
	t.Run("add player to empty roster", func(t *testing.T) {
		r := domain.Roster{
			TeamID:  10,
			Entries: []domain.RosterEntry{},
		}

		err := domain.CanAddPlayer(r, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("add player when already on roster", func(t *testing.T) {
		r := domain.Roster{
			TeamID: 10,
			Entries: []domain.RosterEntry{
				{MLBID: 1},
			},
		}

		err := domain.CanAddPlayer(r, 1)
		if !errors.Is(err, domain.ErrPlayerAlreadyOnRoster) {
			t.Fatalf("got: %v, want: %v", err, domain.ErrPlayerAlreadyOnRoster)
		}
	})

	t.Run("add a second player to roster", func(t *testing.T) {
		r := domain.Roster{
			TeamID: 10,
			Entries: []domain.RosterEntry{
				{MLBID: 1},
			},
		}

		err := domain.CanAddPlayer(r, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
