package domain_test

import (
	"testing"
	"time"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
	"github.com/spcameron/dugout/internal/testsupport/testkit"
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
			effectiveThrough: testkit.TomorrowLock(),
			playerID:         1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
		{
			name:             "allow adding player to roster below cap",
			rosterSize:       domain.MaxRosterSize - 1,
			effectiveThrough: testkit.TomorrowLock(),
			playerID:         domain.MaxRosterSize,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
		{
			name:             "reject adding player to roster at cap",
			rosterSize:       domain.MaxRosterSize,
			effectiveThrough: testkit.TomorrowLock(),
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrRosterFull,
		},
		{
			name:             "reject adding player already on roster",
			rosterSize:       1,
			effectiveThrough: testkit.TomorrowLock(),
			playerID:         1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrPlayerAlreadyOnRoster,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := testkit.NewRosterView(testkit.TeamA(), tc.rosterSize, tc.effectiveThrough)
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
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RoleHitter,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
		{
			name:             "allow activating a pitcher when active pitchers below cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers - 1,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RolePitcher,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
		{
			name:             "reject activating a hitter when active hitters at cap",
			activeHitters:    domain.MaxActiveHitters,
			activePitchers:   0,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RoleHitter,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrActiveHittersFull,
		},
		{
			name:             "reject activating a pitcher when active pitchers at cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RolePitcher,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrActivePitchersFull,
		},
		{
			name:             "allow activating a hitter when active pitchers at cap",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RoleHitter,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
		{
			name:             "allow activating a pitcher when active hitters at cap",
			activeHitters:    domain.MaxActiveHitters,
			activePitchers:   0,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RolePitcher,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          nil,
		},
	}

	for _, tc := range capacityCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize, tc.effectiveThrough),
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
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RoleHitter,
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrPlayerNotOnRoster,
		},
		{
			name:             "reject activating a pitcher not on roster",
			activeHitters:    0,
			activePitchers:   0,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RolePitcher,
			playerID:         domain.MaxRosterSize + 1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrPlayerNotOnRoster,
		},
		{
			name:             "reject activating a hitter when already activated",
			activeHitters:    domain.MaxActiveHitters - 1,
			activePitchers:   0,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RoleHitter,
			playerID:         1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrPlayerAlreadyActive,
		},
		{
			name:             "reject activating a pitcher when already activated",
			activeHitters:    0,
			activePitchers:   domain.MaxActivePitchers - 1,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.RolePitcher,
			playerID:         1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrPlayerAlreadyActive,
		},
		{
			name:             "reject activating player with unknown role",
			activeHitters:    0,
			activePitchers:   0,
			effectiveThrough: testkit.TomorrowLock(),
			role:             domain.PlayerRole(999),
			playerID:         1,
			effectiveAt:      testkit.TomorrowLock(),
			wantErr:          domain.ErrUnrecognizedPlayerRole,
		},
	}

	for _, tc := range membershipCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize, tc.effectiveThrough),
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
			rv := testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), tc.rosterSize, testkit.TodayLock()),
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
			TeamID: testkit.TeamA(),
			Entries: []domain.RosterEntry{
				{
					PlayerID:     1,
					RosterStatus: domain.RosterStatus(999),
				},
			},
		}

		fn := func() { _ = r.Counts() }

		err := require.PanicsError(t, fn)

		require.NotNil(t, err)
		require.ErrorIs(t, err, domain.ErrUnrecognizedRosterStatus)
	})
}

func TestApply(t *testing.T) {
	testAddPlayerCases := []struct {
		name  string
		view  domain.RosterView
		event domain.AddedPlayerToRoster
	}{
		{
			name: "apply AddedPlayerToRoster to empty view creates one inactive entry",
			view: testkit.NewRosterView(testkit.TeamA(), 0, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
		},
		{
			name: "apply AddedPlayerToRoster to view with existing entries appends new inactive entry",
			view: testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize-1, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    domain.MaxRosterSize,
				EffectiveAt: testkit.TodayLock(),
			},
		},
	}

	for _, tc := range testAddPlayerCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.view

			startingTeamID := rv.TeamID
			startingLength := len(rv.Entries)
			startingLock := rv.EffectiveThrough

			var startingLastPlayerID domain.PlayerID
			if startingLength > 0 {
				startingLastPlayerID = rv.Entries[startingLength-1].PlayerID
			}

			rv.Apply(tc.event)

			require.Equal(t, len(rv.Entries), startingLength+1)

			entry := rv.Entries[len(rv.Entries)-1]
			assert.Equal(t, entry.PlayerID, tc.event.PlayerID)
			assert.Equal(t, entry.RosterStatus, domain.StatusInactive)

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
			if startingLength > 0 {
				assert.Equal(t, rv.Entries[startingLength-1].PlayerID, startingLastPlayerID)
			}
		})
	}

	panicCases := []struct {
		name           string
		view           domain.RosterView
		event          domain.AddedPlayerToRoster
		wantErr        error
		wantErrMessage string
	}{
		{
			name: "apply AddedPlayerToRoster panics if TeamID does not match",
			view: testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize-1, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamB(),
				PlayerID:    domain.MaxRosterSize,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrWrongTeamID,
		},
		{
			name: "apply AddedPlayerToRoster panics if event lock outside view effective window",
			view: testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize-1, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    domain.MaxRosterSize,
				EffectiveAt: testkit.TomorrowLock(),
			},
			wantErr: domain.ErrEventOutsideViewWindow,
		},
		{
			name: "apply AddedPlayerToRoster panics if a player with the same ID already present",
			view: testkit.NewRosterView(testkit.TeamA(), 1, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrPlayerAlreadyOnRoster,
		},
	}

	for _, tc := range panicCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.view

			startingTeamID := rv.TeamID
			startingLength := len(rv.Entries)
			startingLock := rv.EffectiveThrough

			fn := func() { rv.Apply(tc.event) }

			if tc.wantErr != nil {
				err := require.PanicsError(t, fn)
				require.NotNil(t, err)
				require.ErrorIs(t, err, tc.wantErr)
			} else if tc.wantErrMessage != "" {
				require.PanicsErrorContains(t, fn, tc.wantErrMessage)
			} else {
				require.Panics(t, fn)
			}

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, len(rv.Entries), startingLength)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
		})
	}
}
