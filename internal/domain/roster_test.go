package domain_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/assert"
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
