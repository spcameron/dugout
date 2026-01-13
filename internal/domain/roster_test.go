package domain_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/require"
)

func TestCanAddPlayer(t *testing.T) {
	t.Run("add player to empty roster", func(t *testing.T) {
		r := roster(100, 0)
		p := domain.Player{
			ID: 1,
		}

		err := domain.CanAddPlayer(r, p.ID)
		require.NoError(t, err)
	})

	t.Run("add a second player to roster", func(t *testing.T) {
		r := roster(100, 1)
		p := domain.Player{
			ID: 2,
		}

		err := domain.CanAddPlayer(r, p.ID)
		require.NoError(t, err)
	})

	t.Run("add player when already on roster", func(t *testing.T) {
		r := roster(100, 1)
		p := domain.Player{
			ID: 1,
		}

		err := domain.CanAddPlayer(r, p.ID)
		require.ErrorIs(t, err, domain.ErrPlayerAlreadyOnRoster)
	})

	t.Run("add a player to 26-man roster", func(t *testing.T) {
		r := roster(100, domain.MaxRosterSize)
		p := domain.Player{
			ID: domain.MaxRosterSize + 1,
		}

		err := domain.CanAddPlayer(r, p.ID)
		require.ErrorIs(t, err, domain.ErrRosterFull)
	})
}

func TestCanActivatePlayer(t *testing.T) {
	t.Run("allow activation when active hitters is below cap", func(t *testing.T) {
		active := domain.MaxActiveHitters - 1
		r := activateRoster(roster(100, domain.MaxRosterSize), active, 0)
		p := domain.Player{
			ID:   domain.PlayerID(active + 1),
			Role: domain.RoleHitter,
		}

		err := domain.CanActivatePlayer(r, p.ID, p.Role)
		require.NoError(t, err)
	})

	t.Run("reject activation when active hitters is at cap", func(t *testing.T) {
		active := domain.MaxActiveHitters
		r := activateRoster(roster(100, domain.MaxRosterSize), active, 0)
		p := domain.Player{
			ID:   domain.PlayerID(active + 1),
			Role: domain.RoleHitter,
		}

		err := domain.CanActivatePlayer(r, p.ID, p.Role)
		require.ErrorIs(t, err, domain.ErrActiveHittersFull)
	})

	t.Run("allow activation when active pitchers is below cap", func(t *testing.T) {
		active := domain.MaxActivePitchers - 1
		r := activateRoster(roster(100, domain.MaxRosterSize), 0, active)
		p := domain.Player{
			ID:   domain.PlayerID(active + 1),
			Role: domain.RolePitcher,
		}

		err := domain.CanActivatePlayer(r, p.ID, p.Role)
		require.NoError(t, err)
	})

	t.Run("reject activation when active pitchers is at cap", func(t *testing.T) {
		active := domain.MaxActivePitchers
		r := activateRoster(roster(100, domain.MaxRosterSize), 0, active)
		p := domain.Player{
			ID:   domain.PlayerID(active + 1),
			Role: domain.RolePitcher,
		}

		err := domain.CanActivatePlayer(r, p.ID, p.Role)
		require.ErrorIs(t, err, domain.ErrActivePitchersFull)
	})

	t.Run("activate a player not on the team", func(t *testing.T) {
		r := roster(100, 1)
		p := domain.Player{
			ID: 2,
		}

		err := domain.CanActivatePlayer(r, p.ID, p.Role)
		require.ErrorIs(t, err, domain.ErrPlayerNotOnRoster)
	})
}

// roster returns a Roster containing a given number of players.
//
// The roster will not exceed the MaxRosterSize, and players will
// be assigned consecutive PlayerIDs beginning from 1.
// Panics if the number of players is less than zero.
func roster(teamID domain.TeamID, players int) domain.Roster {
	if players < 0 {
		panic("players cannot be negative")
	}

	players = min(players, domain.MaxRosterSize)

	r := domain.Roster{
		TeamID:  teamID,
		Entries: make([]domain.RosterEntry, players),
	}

	for i := range players {
		r.Entries[i] = domain.RosterEntry{
			PlayerID:     domain.PlayerID(i + 1),
			RosterStatus: domain.StatusInactive,
		}
	}

	return r
}

// activateRoster returns a Roster with a given number of activate hitters and active pitchers.
//
// The number of hitters and pitchers will not exceed the MaxActiveHitters and MaxActivePitchers.
// Creates a shallow copy of the Roster.Entries slice, so shared ownership is safe.
// Panics if the given number of hitters and pitchers exceeds the length of r.Entries.
func activateRoster(r domain.Roster, hitters, pitchers int) domain.Roster {
	if len(r.Entries) < hitters+pitchers {
		panic("roster entries cannot be fewer than total hitters and pitchers")
	}

	hitters = min(hitters, domain.MaxActiveHitters)
	pitchers = min(pitchers, domain.MaxActivePitchers)

	copyEntries := make([]domain.RosterEntry, len(r.Entries))
	copy(copyEntries, r.Entries)
	r.Entries = copyEntries

	i := 0
	for range hitters {
		r.Entries[i].RosterStatus = domain.StatusActiveHitter
		i++
	}
	for range pitchers {
		r.Entries[i].RosterStatus = domain.StatusActivePitcher
		i++
	}

	return r
}
