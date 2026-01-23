package domain_test

import (
	"testing"
	"time"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/assert"
	"github.com/spcameron/dugout/internal/testutil/require"
)

func TestRosterAppend(t *testing.T) {
	testCases := []struct {
		name       string
		currEvents []domain.RosterEvent
		newEvents  []domain.RosterEvent
		wantErr    error
	}{
		{
			name:       "add one event to empty history",
			currEvents: nil,
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
			},
			wantErr: nil,
		},
		{
			name:       "add multiple events to empty history",
			currEvents: nil,
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    2,
					EffectiveAt: todayLock,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:      teamA,
					PlayerID:    1,
					PlayerRole:  domain.RoleHitter,
					EffectiveAt: todayLock,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:      teamA,
					PlayerID:    2,
					PlayerRole:  domain.RolePitcher,
					EffectiveAt: todayLock,
				},
			},
			wantErr: nil,
		},
		{
			name: "add one event to existing history",
			currEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
			},
			newEvents: []domain.RosterEvent{
				domain.ActivatedPlayerOnRoster{
					TeamID:      teamA,
					PlayerID:    1,
					PlayerRole:  domain.RoleHitter,
					EffectiveAt: todayLock,
				},
			},
			wantErr: nil,
		},
		{
			name: "add multiple events to existing history",
			currEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:      teamA,
					PlayerID:    1,
					PlayerRole:  domain.RoleHitter,
					EffectiveAt: todayLock,
				},
			},
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    2,
					EffectiveAt: todayLock,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:      teamA,
					PlayerID:    2,
					PlayerRole:  domain.RolePitcher,
					EffectiveAt: todayLock,
				},
			},
			wantErr: nil,
		},
		{
			name: "no-op does not change state",
			currEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
			},
			newEvents: nil,
			wantErr:   nil,
		},
		{
			name: "mismatched TeamID throws error and does not mutate",
			currEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
			},
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamB,
					PlayerID:    2,
					EffectiveAt: todayLock,
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := domain.Roster{
				TeamID:       teamA,
				EventHistory: tc.currEvents,
			}

			startingLength := len(r.EventHistory)

			err := r.Append(tc.newEvents...)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(r.EventHistory), startingLength+len(tc.newEvents))
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, len(r.EventHistory), startingLength)
			}
		})
	}
}

func TestRosterProjectThrough(t *testing.T) {
	testCases := []struct {
		name             string
		teamID           domain.TeamID
		effectiveThrough time.Time
		events           []domain.RosterEvent
		wantOnRoster     []domain.PlayerID
		wantOffRoster    []domain.PlayerID
	}{
		{
			name:             "empty history projects to empty view",
			teamID:           teamA,
			effectiveThrough: todayLock,
			events:           nil,
			wantOffRoster:    []domain.PlayerID{1},
		},
		{
			name:             "one event within window projects to view",
			teamID:           teamA,
			effectiveThrough: todayLock,
			events: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "second event outside of window is excluded",
			teamID:           teamA,
			effectiveThrough: todayLock,
			events: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    2,
					EffectiveAt: tomorrowLock,
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "future event in between two past events is not included in view",
			teamID:           teamA,
			effectiveThrough: todayLock,
			events: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    2,
					EffectiveAt: tomorrowLock,
				},
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    3,
					EffectiveAt: todayLock,
				},
			},
			wantOnRoster:  []domain.PlayerID{1, 3},
			wantOffRoster: []domain.PlayerID{2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := domain.Roster{
				TeamID:       tc.teamID,
				EventHistory: tc.events,
			}

			startingHistory := append([]domain.RosterEvent(nil), tc.events...)

			rv := r.ProjectThrough(tc.effectiveThrough)

			assert.Equal(t, rv.TeamID, r.TeamID)
			assert.Equal(t, rv.EffectiveThrough, tc.effectiveThrough)
			assert.Equal(t, r.EventHistory, startingHistory)

			for _, id := range tc.wantOnRoster {
				assert.True(t, rv.PlayerOnRoster(id))
			}

			for _, id := range tc.wantOffRoster {
				assert.False(t, rv.PlayerOnRoster(id))
			}
		})
	}

	panicCases := []struct {
		name             string
		teamID           domain.TeamID
		effectiveThrough time.Time
		events           []domain.RosterEvent
		wantErr          error
	}{
		{
			name:             "panics if event TeamID does not match roster TeamID",
			teamID:           teamA,
			effectiveThrough: todayLock,
			events: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      teamA,
					PlayerID:    1,
					EffectiveAt: todayLock,
				},
				domain.AddedPlayerToRoster{
					TeamID:      teamB,
					PlayerID:    2,
					EffectiveAt: todayLock,
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
	}

	for _, tc := range panicCases {
		t.Run(tc.name, func(t *testing.T) {
			r := domain.Roster{
				TeamID:       tc.teamID,
				EventHistory: tc.events,
			}

			fn := func() { _ = r.ProjectThrough(tc.effectiveThrough) }

			if tc.wantErr != nil {
				err := require.PanicsError(t, fn)
				require.NotNil(t, err)
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}
