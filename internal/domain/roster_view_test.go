package domain_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/assert"
	"github.com/spcameron/dugout/internal/testutil/require"
)

func TestCanAddPlayer(t *testing.T) {
	testCases := []struct {
		name       string
		rosterSize int
		playerID   int
		wantErr    error
	}{
		{
			name:       "allow adding player to empty roster",
			rosterSize: 0,
			playerID:   1,
			wantErr:    nil,
		},
		{
			name:       "allow adding player to roster below cap",
			rosterSize: domain.MaxRosterSize - 1,
			playerID:   domain.MaxRosterSize,
			wantErr:    nil,
		},
		{
			name:       "reject adding player to roster at cap",
			rosterSize: domain.MaxRosterSize,
			playerID:   domain.MaxRosterSize + 1,
			wantErr:    domain.ErrRosterFull,
		},
		{
			name:       "reject adding player already on roster",
			rosterSize: 1,
			playerID:   1,
			wantErr:    domain.ErrPlayerAlreadyOnRoster,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := rosterView(999, tc.rosterSize)
			candidateID := domain.PlayerID(tc.playerID)

			err := r.CanAddPlayer(candidateID)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
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
			r := activatedRosterView(
				rosterView(999, domain.MaxRosterSize),
				tc.activeHitters,
				tc.activePitchers,
			)

			// fixed, known-inactive player
			candidateID := domain.PlayerID(domain.MaxRosterSize)

			err := r.CanActivatePlayer(candidateID, tc.role)

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
		playerID       int
		wantErr        error
	}{
		{
			name:           "reject activating a hitter not on the roster",
			activeHitters:  0,
			activePitchers: 0,
			role:           domain.RoleHitter,
			playerID:       domain.MaxRosterSize + 1,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a pitcher not on the roster",
			activeHitters:  0,
			activePitchers: 0,
			role:           domain.RolePitcher,
			playerID:       domain.MaxRosterSize + 1,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a hitter when already activated",
			activeHitters:  domain.MaxActiveHitters - 1,
			activePitchers: 0,
			role:           domain.RoleHitter,
			playerID:       1,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
		{
			name:           "reject activating a pitcher when already activated",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers - 1,
			role:           domain.RolePitcher,
			playerID:       1,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
	}

	for _, tc := range membershipCases {
		t.Run(tc.name, func(t *testing.T) {
			r := activatedRosterView(
				rosterView(999, domain.MaxRosterSize),
				tc.activeHitters,
				tc.activePitchers,
			)

			candidateID := domain.PlayerID(tc.playerID)

			err := r.CanActivatePlayer(candidateID, tc.role)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestRosterCounts(t *testing.T) {
	testCases := []struct {
		name           string
		rosterSize     int
		activeHitters  int
		activePitchers int
	}{
		{
			name:           "empty roster",
			rosterSize:     0,
			activeHitters:  0,
			activePitchers: 0,
		},
		{
			name:           "full roster with no active hitters or pitchers",
			rosterSize:     domain.MaxRosterSize,
			activeHitters:  0,
			activePitchers: 0,
		},
		{
			name:           "full roster with maximum active hitters and pitchers",
			rosterSize:     domain.MaxRosterSize,
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: domain.MaxActivePitchers,
		},
		{
			name:           "full roster with mid-range active hitters and pitchers",
			rosterSize:     domain.MaxRosterSize,
			activeHitters:  domain.MaxActiveHitters / 2,
			activePitchers: domain.MaxActivePitchers / 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := activatedRosterView(
				rosterView(999, tc.rosterSize),
				tc.activeHitters,
				tc.activePitchers,
			)

			rc := r.Counts()

			assert.Equal(t, rc.Total, tc.rosterSize)
			assert.Equal(t, rc.Total, len(r.Entries))
			assert.Equal(t, rc.ActiveHitters, tc.activeHitters)
			assert.Equal(t, rc.ActivePitchers, tc.activePitchers)
			assert.Equal(t, rc.Inactive, (tc.rosterSize - tc.activeHitters - tc.activePitchers))
		})
	}

	t.Run("panics on unrecognized roster status", func(t *testing.T) {
		r := domain.RosterView{
			TeamID: 999,
			Entries: []domain.RosterEntry{
				{
					PlayerID:     1,
					RosterStatus: domain.RosterStatus(999),
				},
			},
		}

		defer func() {
			got := recover()
			require.NotNil(t, got)

			err, ok := got.(error)
			if !ok {
				t.Fatalf("expected panic value to be error, got %T (%v)", got, got)
			}

			msg := err.Error()
			require.Contains(t, msg, "unrecognized roster status: ")
		}()

		_ = r.Counts()
	})
}

// rosterView returns a Roster containing a given number of players.
//
// The RosterView.Entries will not exceed the MaxRosterSize, and players will be
// assigned consecutive PlayerIDs beginning from 1.
// Panics if the number of players is less than zero.
func rosterView(teamID domain.TeamID, players int) domain.RosterView {
	if players < 0 {
		panic("players cannot be negative")
	}

	players = min(players, domain.MaxRosterSize)

	r := domain.RosterView{
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

// activatedRosterView returns a Roster with a given number of active hitters and active pitchers.
//
// The number of hitters and pitchers will not exceed the MaxActiveHitters and MaxActivePitchers.
// Creates a shallow copy of the RosterView.Entries slice, so shared ownership is safe.
// Panics if the given number of hitters and pitchers exceeds the length of r.Entries.
func activatedRosterView(r domain.RosterView, hitters, pitchers int) domain.RosterView {
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
