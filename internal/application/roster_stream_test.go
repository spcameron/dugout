package application_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/application"
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testsupport/assert"
)

func TestAppend(t *testing.T) {
	testCases := []struct {
		name       string
		currEvents []application.RecordedEvent
		newEvents  []application.RecordedEvent
		wantErr    error
	}{
		{
			name:       "add one event to empty history",
			currEvents: nil,
			newEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:       "add multiple events to empty history",
			currEvents: nil,
			newEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    2,
						EffectiveAt: todayLock,
					},
				},
				{
					Sequence: 3,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      teamA,
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: todayLock,
					},
				},
				{
					Sequence: 4,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      teamA,
						PlayerID:    2,
						PlayerRole:  domain.RolePitcher,
						EffectiveAt: todayLock,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "add one event to existing history",
			currEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
			},
			newEvents: []application.RecordedEvent{
				{
					Sequence: 2,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      teamA,
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: todayLock,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "add multiple event to existing history",
			currEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
				{
					Sequence: 2,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      teamA,
						PlayerID:    1,
						PlayerRole:  domain.RoleHitter,
						EffectiveAt: todayLock,
					},
				},
			},
			newEvents: []application.RecordedEvent{
				{
					Sequence: 3,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    2,
						EffectiveAt: todayLock,
					},
				},
				{
					Sequence: 4,
					Event: domain.ActivatedPlayerOnRoster{
						TeamID:      teamA,
						PlayerID:    2,
						PlayerRole:  domain.RolePitcher,
						EffectiveAt: todayLock,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "no-op does not change state",
			currEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
			},
			newEvents: nil,
			wantErr:   nil,
		},
		{
			name: "mismatched TeamID errors and does not mutate",
			currEvents: []application.RecordedEvent{
				{
					Sequence: 1,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamA,
						PlayerID:    1,
						EffectiveAt: todayLock,
					},
				},
			},
			newEvents: []application.RecordedEvent{
				{
					Sequence: 2,
					Event: domain.AddedPlayerToRoster{
						TeamID:      teamB,
						PlayerID:    2,
						EffectiveAt: todayLock,
					},
				},
			},
			wantErr: domain.ErrWrongTeamID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := application.RosterStream{
				TeamID:       teamA,
				EventHistory: tc.currEvents,
			}

			startingLength := len(rs.EventHistory)

			var startingLastEvent application.RecordedEvent
			if startingLength > 0 {
				startingLastEvent = rs.EventHistory[startingLength-1]
			}

			err := rs.Append(tc.newEvents...)

			if tc.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(rs.EventHistory), startingLength+len(tc.newEvents))
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, len(rs.EventHistory), startingLength)
			}

			if startingLength > 0 {
				assert.Equal(t, rs.EventHistory[startingLength-1], startingLastEvent)
			}
		})
	}
}
