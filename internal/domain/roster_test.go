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
	capacityCases := []struct {
		name           string
		activeHitters  int
		activePitchers int
		role           domain.PlayerRole
		wantErr        error
	}{
		{
			name:           "allow activating a hitter when active hitters is below cap",
			activeHitters:  domain.MaxActiveHitters - 1,
			activePitchers: 0,
			role:           domain.RoleHitter,
			wantErr:        nil,
		},
		{
			name:           "reject activating a hitter when active hitters is at cap",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: 0,
			role:           domain.RoleHitter,
			wantErr:        domain.ErrActiveHittersFull,
		},
		{
			name:           "allow activating a hitter when active pitchers is at cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers,
			role:           domain.RoleHitter,
			wantErr:        nil,
		},
		{
			name:           "allow activating a pitcher when active pitchers is below cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers - 1,
			role:           domain.RolePitcher,
			wantErr:        nil,
		},
		{
			name:           "reject activating a pitcher when active pitchers is at cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers,
			role:           domain.RolePitcher,
			wantErr:        domain.ErrActivePitchersFull,
		},
		{
			name:           "allow activating a pitcher when active hitters is at cap",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: 0,
			role:           domain.RolePitcher,
			wantErr:        nil,
		},
	}

	for _, tc := range capacityCases {
		t.Run(tc.name, func(t *testing.T) {
			r := activateRoster(
				roster(999, domain.MaxRosterSize),
				tc.activeHitters,
				tc.activePitchers,
			)

			// fixed, known-inactive player
			candidateID := domain.PlayerID(domain.MaxRosterSize)

			err := domain.CanActivatePlayer(r, candidateID, tc.role)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}

	membershipCases := []struct {
		name           string
		activeHitters  int
		activePitchers int
		role           domain.PlayerRole
		id             int
		wantErr        error
	}{
		{
			name:           "reject activating a hitter not on the roster",
			activeHitters:  0,
			activePitchers: 0,
			role:           domain.RoleHitter,
			id:             domain.MaxRosterSize + 1,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a pitcher not on the roster",
			activeHitters:  0,
			activePitchers: 0,
			role:           domain.RolePitcher,
			id:             domain.MaxRosterSize + 1,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a hitter when already activated",
			activeHitters:  domain.MaxActiveHitters - 1,
			activePitchers: 0,
			role:           domain.RoleHitter,
			id:             1,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
		{
			name:           "reject activating a pitcher when already activated",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers - 1,
			role:           domain.RolePitcher,
			id:             1,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
	}

	for _, tc := range membershipCases {
		t.Run(tc.name, func(t *testing.T) {
			r := activateRoster(
				roster(999, domain.MaxRosterSize),
				tc.activeHitters,
				tc.activePitchers,
			)

			candidateID := domain.PlayerID(tc.id)

			err := domain.CanActivatePlayer(r, candidateID, tc.role)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

// roster returns a Roster containing a given number of players.
//
// The roster will not exceed the MaxRosterSize, and players will be assigned consecutive PlayerIDs beginning from 1.
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

// activateRoster returns a Roster with a given number of active hitters and active pitchers.
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
