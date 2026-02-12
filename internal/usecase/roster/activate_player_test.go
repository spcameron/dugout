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
		{
			name:       "already activated player returns error and does not append",
			playerID:   1,
			playerRole: domain.RoleHitter,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:   testkit.TeamA(),
					PlayerID: 1,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:     testkit.TeamA(),
					PlayerID:   1,
					PlayerRole: domain.RoleHitter,
				},
			},
			wantErr: domain.ErrPlayerAlreadyActive,
		},
		{
			name:       "player not on roster returns error and does not append",
			playerID:   2,
			playerRole: domain.RolePitcher,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:   testkit.TeamA(),
					PlayerID: 1,
				},
			},
			wantErr: domain.ErrPlayerNotOnRoster,
		},
		{
			name:       "activating pitcher when at max active pitchers returns error and does not append",
			playerID:   domain.MaxActivePitchers + 1,
			playerRole: domain.RolePitcher,
			history:    testkit.GenerateActivatedHistory(testkit.TeamA(), 0, domain.MaxActivePitchers, domain.MaxActivePitchers+1),
			wantErr:    domain.ErrActivePitchersFull,
		},

		{
			name:       "activating hitter when at max active hitters return error and does not append",
			playerID:   domain.MaxActiveHitters + 1,
			playerRole: domain.RoleHitter,
			history:    testkit.GenerateActivatedHistory(testkit.TeamA(), domain.MaxActiveHitters, 0, domain.MaxActiveHitters+1),
			wantErr:    domain.ErrActiveHittersFull,
		},
		{
			name:       "already activated player and mismatched role in ActivatePlayerCommand still errors and does not append",
			playerID:   1,
			playerRole: domain.RoleHitter,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:   testkit.TeamA(),
					PlayerID: 1,
				},
				domain.ActivatedPlayerOnRoster{
					TeamID:     testkit.TeamA(),
					PlayerID:   1,
					PlayerRole: domain.RolePitcher,
				},
			},
			wantErr: domain.ErrPlayerAlreadyActive,
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

				require.Equal(t, len(spy.LoadCalls), 1)
				loadCall := spy.LoadCalls[0]
				assert.Equal(t, loadCall, teamID)

				assert.Equal(t, len(spy.AppendCalls), 0)
			}
		})
	}

	t.Run("load returns error, handle returns error and does not append", func(t *testing.T) {
		store := &testkit.FailingLoadRosterStore{}

		handler := roster.NewActivatePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewActivatePlayerCommand(testkit.TeamA(), 1, domain.RoleHitter)

		err := handler.Handle(cmd)

		assert.ErrorIs(t, err, testkit.ErrFailingLoad)
	})

	t.Run("append returns error, handle returns error", func(t *testing.T) {
		teamID := testkit.TeamA()
		history := testkit.GenerateRosterHistory(teamID, 1)

		store := &testkit.FailingAppendRosterStore{
			Base: testkit.NewFakeRosterStore(),
		}
		store.Base.SeedEvents(teamID, history)

		handler := roster.NewActivatePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewActivatePlayerCommand(teamID, 1, domain.RoleHitter)

		err := handler.Handle(cmd)

		assert.ErrorIs(t, err, testkit.ErrFailingAppend)
	})

	t.Run("append returns ErrVersionConflict, handle returns ErrVersionConflict", func(t *testing.T) {
		teamID := testkit.TeamA()
		history := testkit.GenerateRosterHistory(teamID, 1)

		store := &testkit.VersionConflictRosterStore{
			Base: testkit.NewFakeRosterStore(),
		}
		store.Base.SeedEvents(teamID, history)

		handler := roster.NewActivatePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewActivatePlayerCommand(teamID, 1, domain.RoleHitter)

		err := handler.Handle(cmd)

		assert.ErrorIs(t, err, ports.ErrVersionConflict)
	})
}
