package domain_test

import (
	"slices"
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
	"github.com/spcameron/dugout/internal/testsupport/testkit"
)

func TestDecideAddPlayer(t *testing.T) {
	testCases := []struct {
		name       string
		rosterSize int
		playerID   int
		wantErr    error
	}{
		{
			name:       "accept adding player to empty roster",
			rosterSize: 0,
			playerID:   1,
			wantErr:    nil,
		},
		{
			name:       "accept adding player to roster below cap",
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
			candidateID := domain.PlayerID(tc.playerID)
			rv := testkit.NewRosterView(testkit.TeamA(), tc.rosterSize, testkit.TodayLock())

			events, err := rv.DecideAddPlayer(candidateID)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.AddedPlayerToRoster)
				require.True(t, ok)

				assert.Equal(t, ev.TeamID, testkit.TeamA())
				assert.Equal(t, ev.EffectiveAt, rv.EffectiveThrough)
				assert.Equal(t, ev.PlayerID, candidateID)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestDecideRemovePlayer(t *testing.T) {
	testCases := []struct {
		name       string
		rosterSize int
		playerID   int
		wantErr    error
	}{
		{
			name:       "accept removing player on roster",
			rosterSize: 1,
			playerID:   1,
			wantErr:    nil,
		},
		{
			name:       "reject removing player not on roster",
			rosterSize: 1,
			playerID:   2,
			wantErr:    domain.ErrPlayerNotOnRoster,
		},
		{
			name:       "reject removing player from empty roster",
			rosterSize: 0,
			playerID:   1,
			wantErr:    domain.ErrPlayerNotOnRoster,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			candidateID := domain.PlayerID(tc.playerID)
			rv := testkit.NewRosterView(testkit.TeamA(), tc.rosterSize, testkit.TodayLock())

			events, err := rv.DecideRemovePlayer(candidateID)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.RemovedPlayerFromRoster)
				require.True(t, ok)

				assert.Equal(t, ev.TeamID, testkit.TeamA())
				assert.Equal(t, ev.EffectiveAt, rv.EffectiveThrough)
				assert.Equal(t, ev.PlayerID, candidateID)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestDecideActivatePlayer(t *testing.T) {
	testCases := []struct {
		name           string
		activeHitters  int
		activePitchers int
		playerID       domain.PlayerID
		role           domain.PlayerRole
		wantErr        error
	}{
		{
			name:           "accept activating a hitter when active hitters below cap",
			activeHitters:  domain.MaxActiveHitters - 1,
			activePitchers: 0,
			playerID:       domain.MaxRosterSize,
			role:           domain.RoleHitter,
			wantErr:        nil,
		},
		{
			name:           "accept activating a pitcher when active pitchers below cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers - 1,
			playerID:       domain.MaxRosterSize,
			role:           domain.RolePitcher,
			wantErr:        nil,
		},
		{
			name:           "reject activating a hitter when active hitters at cap",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: 0,
			playerID:       domain.MaxRosterSize,
			role:           domain.RoleHitter,
			wantErr:        domain.ErrActiveHittersFull,
		},
		{
			name:           "reject activating a pitcher when active pitchers at cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers,
			playerID:       domain.MaxRosterSize,
			role:           domain.RolePitcher,
			wantErr:        domain.ErrActivePitchersFull,
		},
		{
			name:           "accept activating a hitter when active pitchers at cap",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers,
			playerID:       domain.MaxRosterSize,
			role:           domain.RoleHitter,
			wantErr:        nil,
		},
		{
			name:           "accept activating a pitcher when active hitters at cap",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: 0,
			playerID:       domain.MaxRosterSize,
			role:           domain.RolePitcher,
			wantErr:        nil,
		},
		{
			name:           "reject activating a hitter not on roster",
			activeHitters:  0,
			activePitchers: 0,
			playerID:       domain.MaxRosterSize + 1,
			role:           domain.RoleHitter,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a pitcher not on roster",
			activeHitters:  0,
			activePitchers: 0,
			playerID:       domain.MaxRosterSize + 1,
			role:           domain.RolePitcher,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject activating a hitter when already activated",
			activeHitters:  domain.MaxActiveHitters - 1,
			activePitchers: 0,
			playerID:       1,
			role:           domain.RoleHitter,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
		{
			name:           "reject activating a pitcher when already activated",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers - 1,
			playerID:       1,
			role:           domain.RolePitcher,
			wantErr:        domain.ErrPlayerAlreadyActive,
		},
		{
			name:           "reject activating player with unknown role",
			activeHitters:  0,
			activePitchers: 0,
			playerID:       1,
			role:           domain.PlayerRole(999),
			wantErr:        domain.ErrUnrecognizedPlayerRole,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			candidateID := domain.PlayerID(tc.playerID)
			rv := testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize, testkit.TodayLock()),
				tc.activeHitters,
				tc.activePitchers,
			)

			events, err := rv.DecideActivatePlayer(candidateID, tc.role)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.ActivatedPlayerOnRoster)
				require.True(t, ok)

				assert.Equal(t, ev.TeamID, testkit.TeamA())
				assert.Equal(t, ev.EffectiveAt, rv.EffectiveThrough)
				assert.Equal(t, ev.PlayerID, candidateID)
				assert.Equal(t, ev.PlayerRole, tc.role)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}

	t.Run("reject activating a player with a bogus roster status", func(t *testing.T) {
		candidateID := domain.PlayerID(1)
		rv := domain.RosterView{
			TeamID: testkit.TeamA(),
			Entries: []domain.RosterEntry{
				{
					TeamID:       testkit.TeamA(),
					PlayerID:     candidateID,
					RosterStatus: domain.RosterStatus(999),
				},
			},
			EffectiveThrough: testkit.TodayLock(),
		}

		events, err := rv.DecideActivatePlayer(candidateID, domain.RoleHitter)

		assert.Nil(t, events)
		assert.ErrorIs(t, err, domain.ErrUnrecognizedRosterStatus)
	})
}

func TestDecideInactivatePlayer(t *testing.T) {
	testCases := []struct {
		name           string
		activeHitters  int
		activePitchers int
		playerID       domain.PlayerID
		wantErr        error
	}{
		{
			name:           "accept inactivating an active hitter on roster",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: 0,
			playerID:       1,
			wantErr:        nil,
		},
		{
			name:           "accept inactivating an active pitcher on roster",
			activeHitters:  0,
			activePitchers: domain.MaxActivePitchers,
			playerID:       1,
			wantErr:        nil,
		},
		{
			name:           "reject inactivating a player not on roster",
			activeHitters:  domain.MaxActiveHitters,
			activePitchers: domain.MaxActivePitchers,
			playerID:       domain.MaxRosterSize + 1,
			wantErr:        domain.ErrPlayerNotOnRoster,
		},
		{
			name:           "reject inactivating a player already inactive",
			activeHitters:  0,
			activePitchers: 0,
			playerID:       1,
			wantErr:        domain.ErrPlayerAlreadyInactive,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize, testkit.TodayLock()),
				tc.activeHitters,
				tc.activePitchers,
			)

			candidateID := domain.PlayerID(tc.playerID)

			events, err := rv.DecideInactivatePlayer(candidateID)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				require.Equal(t, len(events), 1)

				ev, ok := events[0].(domain.InactivatedPlayerOnRoster)
				require.True(t, ok)

				assert.Equal(t, ev.TeamID, testkit.TeamA())
				assert.Equal(t, ev.EffectiveAt, rv.EffectiveThrough)
				assert.Equal(t, ev.PlayerID, candidateID)
			} else {
				assert.Nil(t, events)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}

	t.Run("reject inactivating a player with a bogus roster status", func(t *testing.T) {
		candidateID := domain.PlayerID(1)
		rv := domain.RosterView{
			TeamID: testkit.TeamA(),
			Entries: []domain.RosterEntry{
				{
					TeamID:       testkit.TeamA(),
					PlayerID:     candidateID,
					RosterStatus: domain.RosterStatus(999),
				},
			},
			EffectiveThrough: testkit.TodayLock(),
		}

		events, err := rv.DecideInactivatePlayer(candidateID)

		assert.Nil(t, events)
		assert.ErrorIs(t, err, domain.ErrUnrecognizedRosterStatus)
	})
}

func TestCounts(t *testing.T) {
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
			name: "adding player to empty view creates one inactive entry",
			view: testkit.NewRosterView(testkit.TeamA(), 0, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
		},
		{
			name: "adding player to view with existing entries appends new inactive entry",
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
			assert.Equal(t, entry.TeamID, tc.event.TeamID)
			assert.Equal(t, entry.PlayerID, tc.event.PlayerID)
			assert.Equal(t, entry.RosterStatus, domain.StatusInactive)

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
			if startingLength > 0 {
				assert.Equal(t, rv.Entries[startingLength-1].PlayerID, startingLastPlayerID)
			}
		})
	}

	testRemovePlayerCases := []struct {
		name          string
		view          domain.RosterView
		event         domain.RemovedPlayerFromRoster
		wasPresent    bool
		wantOnRoster  []domain.PlayerID
		wantOffRoster []domain.PlayerID
	}{
		{
			name: "removing player from empty view is no-op",
			view: testkit.NewRosterView(testkit.TeamA(), 0, testkit.TodayLock()),
			event: domain.RemovedPlayerFromRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
			wasPresent:    false,
			wantOffRoster: []domain.PlayerID{1},
		},
		{
			name: "removing player not on roster is no-op",
			view: testkit.NewRosterView(testkit.TeamA(), 1, testkit.TodayLock()),
			event: domain.RemovedPlayerFromRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    2,
				EffectiveAt: testkit.TodayLock(),
			},
			wasPresent:    false,
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name: "removing a player on roster deletes the roster entry",
			view: testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
			event: domain.RemovedPlayerFromRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    2,
				EffectiveAt: testkit.TodayLock(),
			},
			wasPresent:    true,
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name: "removing a player on roster does not affect the entries before or after",
			view: testkit.NewRosterView(testkit.TeamA(), 3, testkit.TodayLock()),
			event: domain.RemovedPlayerFromRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    2,
				EffectiveAt: testkit.TodayLock(),
			},
			wasPresent:    true,
			wantOnRoster:  []domain.PlayerID{1, 3},
			wantOffRoster: []domain.PlayerID{2},
		},
	}

	for _, tc := range testRemovePlayerCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.view

			startingTeamID := rv.TeamID
			startingLength := len(rv.Entries)
			startingLock := rv.EffectiveThrough
			startingEntries := slices.Clone(rv.Entries)

			rv.Apply(tc.event)

			if tc.wasPresent {
				require.Equal(t, len(rv.Entries), startingLength-1)
			} else {
				require.Equal(t, len(rv.Entries), startingLength)
				require.Equal(t, rv.Entries, startingEntries)
			}

			for _, id := range tc.wantOnRoster {
				assert.True(t, rv.PlayerOnRoster(id))
			}

			for _, id := range tc.wantOffRoster {
				assert.False(t, rv.PlayerOnRoster(id))
			}

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
		})
	}

	testActivatePlayerCases := []struct {
		name         string
		view         domain.RosterView
		event        domain.RosterEvent
		wantHitters  []domain.PlayerID
		wantPitchers []domain.PlayerID
		wantInactive []domain.PlayerID
	}{
		{
			name: "activating a hitter on roster updates entry to active hitter status",
			view: testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				PlayerRole:  domain.RoleHitter,
				EffectiveAt: testkit.TodayLock(),
			},
			wantHitters:  []domain.PlayerID{1},
			wantInactive: []domain.PlayerID{2},
		},
		{
			name: "activating a pitcher on roster updates entry to active pitcher status",
			view: testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				PlayerRole:  domain.RolePitcher,
				EffectiveAt: testkit.TodayLock(),
			},
			wantPitchers: []domain.PlayerID{1},
			wantInactive: []domain.PlayerID{2},
		},
		{
			name: "activating a hitter already with active hitter status is no-op",
			view: testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
				1, 0,
			),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				PlayerRole:  domain.RoleHitter,
				EffectiveAt: testkit.TodayLock(),
			},
			wantHitters:  []domain.PlayerID{1},
			wantInactive: []domain.PlayerID{2},
		},
		{
			name: "activating a pitcher already with active pitcher status is no-op",
			view: testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
				0, 1,
			),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				PlayerRole:  domain.RolePitcher,
				EffectiveAt: testkit.TodayLock(),
			},
			wantPitchers: []domain.PlayerID{1},
			wantInactive: []domain.PlayerID{2},
		},
	}

	for _, tc := range testActivatePlayerCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.view

			startingTeamID := rv.TeamID
			startingLength := len(rv.Entries)
			startingLock := rv.EffectiveThrough

			rv.Apply(tc.event)

			for _, e := range rv.Entries {
				id := e.PlayerID
				status := e.RosterStatus

				switch {
				case slices.Contains(tc.wantHitters, id):
					assert.Equal(t, status, domain.StatusActiveHitter)
				case slices.Contains(tc.wantPitchers, id):
					assert.Equal(t, status, domain.StatusActivePitcher)
				case slices.Contains(tc.wantInactive, id):
					assert.Equal(t, status, domain.StatusInactive)
				}

				assert.Equal(t, e.TeamID, rv.TeamID)
			}

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, len(rv.Entries), startingLength)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
		})
	}

	testInactivatePlayerCases := []struct {
		name         string
		view         domain.RosterView
		event        domain.RosterEvent
		wantHitters  []domain.PlayerID
		wantPitchers []domain.PlayerID
		wantInactive []domain.PlayerID
	}{
		{
			name: "inactivating an active player on roster updates entry to inactive status",
			view: testkit.ActivatedRosterView(
				testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
				2, 0,
			),
			event: domain.InactivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
			wantHitters:  []domain.PlayerID{2},
			wantInactive: []domain.PlayerID{1},
		},
		{
			name: "inactivating an inactive player on roster is no-op",
			view: testkit.NewRosterView(testkit.TeamA(), 2, testkit.TodayLock()),
			event: domain.InactivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
			wantInactive: []domain.PlayerID{1, 2},
		},
	}

	for _, tc := range testInactivatePlayerCases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.view

			startingTeamID := rv.TeamID
			startingLength := len(rv.Entries)
			startingLock := rv.EffectiveThrough

			rv.Apply(tc.event)

			for _, e := range rv.Entries {
				id := e.PlayerID
				status := e.RosterStatus

				switch {
				case slices.Contains(tc.wantHitters, id):
					assert.Equal(t, status, domain.StatusActiveHitter)
				case slices.Contains(tc.wantPitchers, id):
					assert.Equal(t, status, domain.StatusActivePitcher)
				case slices.Contains(tc.wantInactive, id):
					assert.Equal(t, status, domain.StatusInactive)
				}

				assert.Equal(t, e.TeamID, rv.TeamID)
			}

			assert.Equal(t, rv.TeamID, startingTeamID)
			assert.Equal(t, len(rv.Entries), startingLength)
			assert.Equal(t, rv.EffectiveThrough, startingLock)
		})
	}

	panicCases := []struct {
		name           string
		view           domain.RosterView
		event          domain.RosterEvent
		wantErr        error
		wantErrMessage string
	}{
		{
			name: "apply RosterEvent panics if TeamID does not match",
			view: testkit.NewRosterView(testkit.TeamA(), domain.MaxRosterSize-1, testkit.TodayLock()),
			event: domain.AddedPlayerToRoster{
				TeamID:      testkit.TeamB(),
				PlayerID:    domain.MaxRosterSize,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrWrongTeamID,
		},
		{
			name: "apply RosterEvent panics if event lock outside view effective window",
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
		{
			name: "apply ActivatedPlayerOnRoster panics if player role is unrecognized",
			view: testkit.NewRosterView(testkit.TeamA(), 1, testkit.TodayLock()),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				PlayerRole:  domain.PlayerRole(999),
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrUnrecognizedPlayerRole,
		},
		{
			name: "apply ActivatedPlayerOnRoster panics if player is not already on roster",
			view: testkit.NewRosterView(testkit.TeamA(), 1, testkit.TodayLock()),
			event: domain.ActivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    2,
				PlayerRole:  domain.RoleHitter,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrPlayerNotOnRoster,
		},
		{
			name: "apply InactivatedPlayerOnRoster panics if current roster status is unrecognized",
			view: domain.RosterView{

				TeamID: testkit.TeamA(),
				Entries: []domain.RosterEntry{
					{
						TeamID:       testkit.TeamA(),
						PlayerID:     1,
						RosterStatus: domain.RosterStatus(999),
					},
				},
				EffectiveThrough: testkit.TodayLock(),
			},
			event: domain.InactivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    1,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrUnrecognizedRosterStatus,
		},
		{
			name: "apply InactivatedPlayerOnRoster panics if player is not already on roster",
			view: testkit.NewRosterView(testkit.TeamA(), 1, testkit.TodayLock()),
			event: domain.InactivatedPlayerOnRoster{
				TeamID:      testkit.TeamA(),
				PlayerID:    2,
				EffectiveAt: testkit.TodayLock(),
			},
			wantErr: domain.ErrPlayerNotOnRoster,
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
