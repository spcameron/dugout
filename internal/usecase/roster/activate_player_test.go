package roster_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/ports"
	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
	"github.com/spcameron/dugout/internal/testsupport/testkit"
	"github.com/spcameron/dugout/internal/usecase/roster"
)

func TestActivatePlayerHandler_Handle(t *testing.T) {
	testCases := []struct {
		name       string
		playerID   domain.PlayerID
		playerRole domain.PlayerRole
		history    []domain.RosterEvent
		wantErr    error
	}{
		{
			name:       "inactive pitcher on roster appends ActivatedPlayerOnRoster event",
			playerID:   1,
			playerRole: domain.RolePitcher,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:   testkit.TeamA(),
					PlayerID: 1,
				},
			},
			wantErr: nil,
		},
		{
			name:       "inactive hitter on roster appends ActivatedPlayerOnRoster event",
			playerID:   1,
			playerRole: domain.RoleHitter,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:   testkit.TeamA(),
					PlayerID: 1,
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			teamID := testkit.TeamA()

			leagueLock := testkit.NewStubLeagueLock()
			store := testkit.NewFakeRosterStore()
			spy := testkit.NewSpyRosterStore(store)

			store.SeedEvents(teamID, tc.history)

			handler := roster.NewActivatePlayerHandler(spy, leagueLock)
			cmd := roster.NewActivatePlayerCommand(teamID, tc.playerID, tc.playerRole)

			err := handler.Handle(cmd)

			if tc.wantErr == nil {
				assert.Nil(t, err)

				require.Equal(t, len(spy.LoadCalls), 1)
				loadCall := spy.LoadCalls[0]
				assert.Equal(t, loadCall, teamID)

				require.Equal(t, len(spy.AppendCalls), 1)
				appendCall := spy.AppendCalls[0]
				assert.Equal(t, appendCall.TeamID, teamID)
				assert.Equal(t, appendCall.Version, ports.Version(len(tc.history)))

				require.Equal(t, len(appendCall.Events), 1)
				appendedEvent := appendCall.Events[0]
				require.Equal(t, appendedEvent.Team(), teamID)
				require.Equal(t, appendedEvent.OccurredAt(), handler.Lock.NextLock())

				ev, ok := appendedEvent.(domain.ActivatedPlayerOnRoster)
				require.True(t, ok)
				assert.Equal(t, ev.PlayerID, tc.playerID)
				assert.Equal(t, ev.PlayerRole, tc.playerRole)

			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}
