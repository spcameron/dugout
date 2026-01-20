package domain_test

import (
	"testing"
	"time"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/assert"
	"github.com/spcameron/dugout/internal/testutil/require"
)

func TestDecideAddPlayer(t *testing.T) {
	testCases := []struct {
		name             string
		rosterSize       int
		effectiveThrough time.Time
		playerID         int
		effectiveAt      time.Time
		wantErr          error
	}{
		{
			name:             "allow adding player to empty roster",
			rosterSize:       0,
			effectiveThrough: tomorrowLock,
			playerID:         1,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
		{
			name:             "allow adding player to roster below cap",
			rosterSize:       domain.MaxRosterSize - 1,
			effectiveThrough: tomorrowLock,
			playerID:         domain.MaxRosterSize,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
		{
			name:             "reject adding player to roster at cap",
			rosterSize:       domain.MaxRosterSize,
			effectiveThrough: tomorrowLock,
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrRosterFull,
		},
		{
			name:             "reject adding player already on roster",
			rosterSize:       1,
			effectiveThrough: tomorrowLock,
			playerID:         1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrPlayerAlreadyOnRoster,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := rosterView(teamA, tc.rosterSize, tc.effectiveThrough)
			candidateID := domain.PlayerID(tc.playerID)

			events, err := rv.DecideAddPlayer(candidateID, tc.effectiveAt)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.AddedPlayerToRoster)
				require.True(t, ok)

				assert.Equal(t, ev.EffectiveAt, tc.effectiveAt)
				assert.Equal(t, ev.PlayerID, candidateID)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestDecideActivatePlayer(t *testing.T) {
	capacityCases := []struct {
		name             string
		activeHitters    int
		activePitchers   int
		effectiveThrough time.Time
		role             domain.PlayerRole
		effectiveAt      time.Time
		wantErr          error
	}{
		{
			name:             "allow activating a hitter when active hitters below cap",
			activeHitters:    domain.MaxActiveHitters - 1,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RoleHitter,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
		{
			name:             "allow activating a pitcher when active pitchers below cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers - 1,
			effectiveThrough: tomorrowLock,
			role:             domain.RolePitcher,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
		{
			name:             "reject activating a hitter when active hitters at cap",
			activeHitters:    domain.MaxActiveHitters,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RoleHitter,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrActiveHittersFull,
		},
		{
			name:             "reject activating a pitcher when active pitchers at cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers,
			effectiveThrough: tomorrowLock,
			role:             domain.RolePitcher,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrActivePitchersFull,
		},
		{
			name:             "allow activating a hitter when active pitchers at cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers,
			effectiveThrough: tomorrowLock,
			role:             domain.RoleHitter,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
		{
			name:             "allow activating a pitcher when active hitters at cap",
			activeHitters:    domain.MaxActiveHitters,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RolePitcher,
			effectiveAt:      tomorrowLock,
			wantErr:          nil,
		},
	}

	for _, tc := range capacityCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := activatedRosterView(
				rosterView(teamA, domain.MaxRosterSize, tc.effectiveThrough),
				tc.activeHitters,
				tc.activePitchers,
			)

			// fixed, known-inactive player
			candidateID := domain.PlayerID(domain.MaxRosterSize)

			events, err := rv.DecideActivatePlayer(candidateID, tc.role, tc.effectiveAt)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.ActivatedPlayerOnRoster)
				require.True(t, ok)

				assert.Equal(t, ev.EffectiveAt, tc.effectiveAt)
				assert.Equal(t, ev.PlayerID, candidateID)
				assert.Equal(t, ev.PlayerRole, tc.role)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}

	membershipCases := []struct {
		name             string
		activeHitters    int
		activePitchers   int
		effectiveThrough time.Time
		role             domain.PlayerRole
		playerID         int
		effectiveAt      time.Time
		wantErr          error
	}{
		{
			name:             "reject activating a hitter not on roster",
			activeHitters:    0,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RoleHitter,
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrPlayerNotOnRoster,
		},
		{
			name:             "reject activating a pitcher not on roster",
			activeHitters:    0,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RolePitcher,
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrPlayerNotOnRoster,
		},
		{
			name:             "reject activating a hitter when already activated",
			activeHitters:    domain.MaxActiveHitters - 1,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.RoleHitter,
			playerID:         1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrPlayerAlreadyActive,
		},
		{
			name:             "reject activating a pitcher when already activated",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers - 1,
			effectiveThrough: tomorrowLock,
			role:             domain.RolePitcher,
			playerID:         1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrPlayerAlreadyActive,
		},
		{
			name:             "reject activating player with unknown role",
			activeHitters:    0,
			activePitchers:   0,
			effectiveThrough: tomorrowLock,
			role:             domain.PlayerRole(999),
			playerID:         1,
			effectiveAt:      tomorrowLock,
			wantErr:          domain.ErrUnrecognizedPlayerRole,
		},
	}

	for _, tc := range membershipCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := activatedRosterView(
				rosterView(teamA, domain.MaxRosterSize, tc.effectiveThrough),
				tc.activeHitters,
				tc.activePitchers,
			)

			candidateID := domain.PlayerID(tc.playerID)

			events, err := rv.DecideActivatePlayer(candidateID, tc.role, tc.effectiveAt)

			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
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
			rv := activatedRosterView(
				rosterView(teamA, tc.rosterSize, todayLock),
				tc.activeHitters,
				tc.activePitchers,
			)

			rc := rv.Counts()

			assert.Equal(t, rc.Total, tc.rosterSize)
			assert.Equal(t, rc.Total, len(rv.Entries))
			assert.Equal(t, rc.ActiveHitters, tc.activeHitters)
			assert.Equal(t, rc.ActivePitchers, tc.activePitchers)
			assert.Equal(t, rc.Inactive, (tc.rosterSize - tc.activeHitters - tc.activePitchers))
		})
	}

	t.Run("panics on unrecognized roster status", func(t *testing.T) {
		r := domain.RosterView{
			TeamID: teamA,
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
			assert.Contains(t, msg, "unrecognized roster status: ")
		}()

		_ = r.Counts()
	})
}

// rosterView returns a Roster containing a given number of players.
//
// Players will be assigned consecutive PlayerIDs beginning from 1.
// Panics if the number of players is less than zero, or greater than MaxRosterSize.
func rosterView(teamID domain.TeamID, players int, cutoff time.Time) domain.RosterView {
	if players < 0 {
		panic("players cannot be negative")
	}

	if players > domain.MaxRosterSize {
		panic("players exceeds MaxRosterSize")
	}

	rv := domain.RosterView{
		TeamID:           teamID,
		Entries:          make([]domain.RosterEntry, players),
		EffectiveThrough: cutoff,
	}

	for i := range players {
		rv.Entries[i] = domain.RosterEntry{
			PlayerID:     domain.PlayerID(i + 1),
			RosterStatus: domain.StatusInactive,
		}
	}

	return rv
}

// activatedRosterView returns a Roster with a given number of active hitters and active pitchers.
//
// The number of hitters and pitchers will not exceed the MaxActiveHitters and MaxActivePitchers.
// Creates a shallow copy of the RosterView.Entries slice, so shared ownership is safe.
// Panics if the given number of hitters and pitchers exceeds the length of r.Entries.
func activatedRosterView(rv domain.RosterView, hitters, pitchers int) domain.RosterView {
	if len(rv.Entries) < hitters+pitchers {
		panic("roster entries cannot be fewer than total hitters and pitchers")
	}

	hitters = min(hitters, domain.MaxActiveHitters)
	pitchers = min(pitchers, domain.MaxActivePitchers)

	copyEntries := make([]domain.RosterEntry, len(rv.Entries))
	copy(copyEntries, rv.Entries)
	rv.Entries = copyEntries

	i := 0
	for range hitters {
		rv.Entries[i].RosterStatus = domain.StatusActiveHitter
		i++
	}
	for range pitchers {
		rv.Entries[i].RosterStatus = domain.StatusActivePitcher
		i++
	}

	return rv
}
