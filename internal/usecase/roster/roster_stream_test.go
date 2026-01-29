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

func TestStage(t *testing.T) {
	testCases := []struct {
		name            string
		committedEvents []roster.RecordedRosterEvent
		newEvents       []domain.RosterEvent
		wantErr         error
	}{
		{
			name:            "add one event to empty pending events",
			committedEvents: nil,
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantErr: nil,
		},
		{
			name:            "add multiple events to empty pending events",
			committedEvents: nil,
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					PlayerRole:  domain.RoleHitter,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantErr: nil,
		},
		{
			name: "no-op leaves pending unchanged",
			committedEvents: []roster.RecordedRosterEvent{
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
			committedEvents: []roster.RecordedRosterEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      testkit.TeamA(),
						PlayerID:    1,
						EffectiveAt: testkit.TodayLock(),
					},
				},
			},
			newEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamB(),
					PlayerID:    2,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := roster.RosterStream{
				TeamID:    testkit.TeamA(),
				Committed: tc.committedEvents,
			}

			startingCommittedLength := len(rs.Committed)
			startingPendingLength := len(rs.Pending)

			err := rs.Stage(tc.newEvents...)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(rs.Pending), startingPendingLength+len(tc.newEvents))
				assert.Equal(t, rs.Pending[startingPendingLength:], tc.newEvents)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, len(rs.Pending), startingPendingLength)
			}

			assert.Equal(t, len(rs.Committed), startingCommittedLength)
		})
	}
}

func TestProjectThrough(t *testing.T) {
	testCases := []struct {
		name             string
		teamID           domain.TeamID
		effectiveThrough time.Time
		committedEvents  []roster.RecordedRosterEvent
		pendingEvents    []domain.RosterEvent
		wantOnRoster     []domain.PlayerID
		wantOffRoster    []domain.PlayerID
	}{
		{
			name:             "empty committed and empty pending projects to empty view",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			committedEvents:  nil,
			pendingEvents:    nil,
			wantOnRoster:     nil,
		},
		{
			name:             "committed events are filtered by effectiveThrough cutoff time",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			committedEvents: []roster.RecordedRosterEvent{
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
			name:             "committed events are applied in sequence order, not input order",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			committedEvents: []roster.RecordedRosterEvent{
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
			name:             "pending events are filtered by effectiveThrough cutoff time",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			pendingEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    2,
					EffectiveAt: testkit.TomorrowLock(),
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: []domain.PlayerID{2},
		},
		{
			name:             "pending preserves staging order (add then remove)",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			pendingEvents: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
				domain.RemovedPlayerFromRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantOnRoster:  nil,
			wantOffRoster: []domain.PlayerID{1},
		},
		{
			name:             "pending preserves staging order (remove then add)",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			pendingEvents: []domain.RosterEvent{
				domain.RemovedPlayerFromRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantOnRoster:  []domain.PlayerID{1},
			wantOffRoster: nil,
		},
		{
			name:             "committed is always applied before pending",
			teamID:           testkit.TeamA(),
			effectiveThrough: testkit.TodayLock(),
			committedEvents: []roster.RecordedRosterEvent{
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
			},
			pendingEvents: []domain.RosterEvent{
				domain.RemovedPlayerFromRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantOnRoster:  []domain.PlayerID{2},
			wantOffRoster: []domain.PlayerID{1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := roster.RosterStream{
				TeamID:    tc.teamID,
				Committed: tc.committedEvents,
				Pending:   tc.pendingEvents,
			}

			startingCommittedEvents := slices.Clone(tc.committedEvents)
			startingPendingEvents := slices.Clone(tc.pendingEvents)

			rv := rs.ProjectThrough(tc.effectiveThrough)

			assert.Equal(t, rv.TeamID, rs.TeamID)
			assert.Equal(t, rv.EffectiveThrough, tc.effectiveThrough)
			assert.Equal(t, rs.Committed, startingCommittedEvents)
			assert.Equal(t, rs.Pending, startingPendingEvents)

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
				TeamID:    tc.teamID,
				Committed: tc.events,
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
