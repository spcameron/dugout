package roster_test

import (
	"slices"
	"testing"
	"time"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
	"github.com/spcameron/dugout/internal/testsupport/testkit"
	"github.com/spcameron/dugout/internal/usecase/roster"
)

func TestAppend(t *testing.T) {
	testCases := []struct {
		name       string
		currEvents []roster.RecordedRosterEvent
		newEvents  []roster.RecordedRosterEvent
		wantErr    error
	}{
		{
			name:       "add one event to empty history",
			currEvents: nil,
			newEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name:       "add multiple events to empty history",
			currEvents: nil,
			newEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 3,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 4,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						PlayerRole:  domain.RolePitcher,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "add one event to existing history",
			currEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			newEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 2,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "add multiple event to existing history",
			currEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			newEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 3,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 4,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						PlayerRole:  domain.RolePitcher,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "no-op does not change state",
			currEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			newEvents: nil,
			wantErr:   nil,
		},
		{
			name: "mismatched TeamID errors and does not mutate",
			currEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			newEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamB(),
						PlayerID:    2,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := roster.RosterStream{
				TeamID:         testkit.TeamA(),
				RecordedEvents: tc.currEvents,
			}

			startingLength := len(rs.RecordedEvents)

			var startingLastEvent roster.RecordedRosterEvent
			if startingLength > 0 {
				startingLastEvent = rs.RecordedEvents[startingLength-1]
			}

			err := rs.Append(tc.newEvents...)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(rs.RecordedEvents), startingLength+len(tc.newEvents))
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, len(rs.RecordedEvents), startingLength)
			}

			if startingLength > 0 {
				assert.Equal(t, rs.RecordedEvents[startingLength-1], startingLastEvent)
			}
		})
	}
}

func TestProjectThrough(t *testing.T) {
	testCases := []struct {
		name             string
		teamID           domain.TeamID
		effectiveThrough time.Time
		events           []roster.RecordedRosterEvent
		wantOnRoster     []domain.PlayerID
		wantOffRoster    []domain.PlayerID
	}{
		{
			name:             "empty history projects to empty view",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events:           nil,
			wantOnRoster:     nil,
		},
		{
			name:             "one event within window projects to view",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "second event outside of window is excluded",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						EffectiveAt: testkit.TomorrowLock(),
					},
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "future event in between two past events is not included in view",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    2,
						EffectiveAt: testkit.TomorrowLock(),
					},
				},
				{
					Sequence: 3,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    3,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantOnRoster:  []domain.PlayerID{1, 3},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "applies event in sequence order, not input order",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 2,
					Event: domain.RemovedPlayerFromRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantOnRoster:  nil,
			wantOffRoster: []domain.PlayerID{1},
		},
		{
			name:             "filtering events by time happens independent of sorting",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.RemovedPlayerFromRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TomorrowLock(),
					},
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := roster.RosterStream{
				TeamID:         tc.teamID,
				RecordedEvents: tc.events,
			}

			startingHistory := slices.Clone(tc.events)

			rv := rs.ProjectThrough(tc.effectiveThrough)

			assert.Equal(t, rv.TeamID, rs.TeamID)
			assert.Equal(t, rv.EffectiveThrough, tc.effectiveThrough)
			assert.Equal(t, rs.RecordedEvents, startingHistory)

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
		events           []roster.RecordedRosterEvent
		wantErr          error
	}{
		{
			name:             "panics if event TeamID does not match roster stream TeamID",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamB(),
						PlayerID:    2,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
		{
			name:             "panics if two events have the same sequence value",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			events: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
				{
					Sequence: 1,
					Event: domain.RemovedPlayerFromRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			wantErr: roster.ErrDuplicateRecordedEventSequence,
		},
	}

	for _, tc := range panicCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := roster.RosterStream{
				TeamID:         tc.teamID,
				RecordedEvents: tc.events,
			}

			fn := func() { _ = rs.ProjectThrough(tc.effectiveThrough) }

			if tc.wantErr != nil {
				err := require.PanicsError(t, fn)
				require.NotNil(t, err)
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.Panics(t, fn)
			}

		})
	}
}
