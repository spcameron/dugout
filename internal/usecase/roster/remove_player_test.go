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

func TestRemovePlayerHandler_Handle(t *testing.T) {
	testCases := []struct {
		name     string
		playerID domain.PlayerID
		history  []domain.RosterEvent
		wantErr  error
	}{
		{
			name:     "player on roster appends RemovePlayerFromRoster event",
			playerID: 1,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
		},
		{
			name:     "empty history returns error and does not append",
			playerID: 1,
			history:  nil,
			wantErr:  domain.ErrPlayerNotOnRoster,
		},
		{
			name:     "non-empty history & player not on roster returns error and does not append",
			playerID: 2,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantErr: domain.ErrPlayerNotOnRoster,
		},
		{
			name:     "player already removed from roster returns error and does not append",
			playerID: 1,
			history: []domain.RosterEvent{
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
			wantErr: domain.ErrPlayerNotOnRoster,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			teamID := testkit.TeamA()

			leagueLock := testkit.NewStubLeagueLock()
			store := testkit.NewFakeRosterStore()
			spy := testkit.NewSpyRosterStore(store)

			store.SeedEvents(teamID, tc.history)

			handler := roster.NewRemovePlayerHandler(spy, leagueLock)
			cmd := roster.NewRemovePlayerCommand(teamID, tc.playerID)

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

				ev, ok := appendedEvent.(domain.RemovedPlayerFromRoster)
				require.True(t, ok)
				assert.Equal(t, ev.PlayerID, tc.playerID)
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

		handler := roster.NewRemovePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewRemovePlayerCommand(testkit.TeamA(), 1)

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

		handler := roster.NewRemovePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewRemovePlayerCommand(teamID, 1)

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

		handler := roster.NewRemovePlayerHandler(store, testkit.NewStubLeagueLock())
		cmd := roster.NewRemovePlayerCommand(teamID, 1)

		err := handler.Handle(cmd)

		assert.ErrorIs(t, err, ports.ErrVersionConflict)
	})
}
