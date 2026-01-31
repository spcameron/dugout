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

func TestAddPlayerHandler_Handle(t *testing.T) {
	testCases := []struct {
		name     string
		teamID   domain.TeamID
		playerID domain.PlayerID
		history  []domain.RosterEvent
		wantErr  error
	}{
		{
			name:     "empty history appends AddPlayerToRoster event",
			teamID:   testkit.TeamA(),
			playerID: 1,
			history:  nil,
			wantErr:  nil,
		},
		{
			name:     "player already on projected roster returns error and does not append",
			teamID:   testkit.TeamA(),
			playerID: 1,
			history: []domain.RosterEvent{
				domain.AddedPlayerToRoster{
					TeamID:      testkit.TeamA(),
					PlayerID:    1,
					EffectiveAt: testkit.TodayLock(),
				},
			},
			wantErr: domain.ErrPlayerAlreadyOnRoster,
		},
		{
			name:     "adding player to full roster returns error and does not append",
			teamID:   testkit.TeamA(),
			playerID: domain.MaxRosterSize + 1,
			history:  generateRosterHistory(testkit.TeamA(), domain.MaxRosterSize),
			wantErr:  domain.ErrRosterFull,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			leagueLock := testkit.NewStubLeagueLock()
			store := testkit.NewFakeRosterStore()
			spy := testkit.NewSpyRosterStore(store)

			store.SeedEvents(tc.teamID, tc.history)

			handler := roster.NewAddPlayerHandler(spy, leagueLock)
			cmd := roster.NewAddPlayerCommand(tc.teamID, tc.playerID)

			err := handler.Handle(cmd)

			if tc.wantErr == nil {
				assert.Nil(t, err)

				require.Equal(t, len(spy.LoadCalls), 1)
				loadCall := spy.LoadCalls[0]
				assert.Equal(t, loadCall, tc.teamID)

				require.Equal(t, len(spy.AppendCalls), 1)
				appendCall := spy.AppendCalls[0]
				assert.Equal(t, appendCall.TeamID, tc.teamID)
				assert.Equal(t, appendCall.Version, ports.Version(len(tc.history)))

				require.Equal(t, len(appendCall.Events), 1)
				appendedEvent := appendCall.Events[0]
				require.Equal(t, appendedEvent.Team(), tc.teamID)
				require.Equal(t, appendedEvent.OccurredAt(), handler.Lock.NextLock())

				ev, ok := appendedEvent.(domain.AddedPlayerToRoster)
				require.True(t, ok)
				assert.Equal(t, ev.PlayerID, tc.playerID)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)

				require.Equal(t, len(spy.LoadCalls), 1)
				loadCall := spy.LoadCalls[0]
				assert.Equal(t, loadCall, tc.teamID)

				assert.Equal(t, len(spy.AppendCalls), 0)
			}
		})
	}

	failureTestCases := []struct {
		name    string
		store   ports.RosterStore
		wantErr error
	}{
		{
			name:    "load returns error, handle returns error and does not append",
			store:   &testkit.FailingLoadRosterStore{},
			wantErr: testkit.ErrFailingLoad,
		},
		{
			name:    "append returns error, handle returns error",
			store:   &testkit.FailingAppendRosterStore{},
			wantErr: testkit.ErrFailingAppend,
		},
		{
			name:    "append return ErrVersionConflict, handle returns ErrVersionConflict",
			store:   &testkit.VersionConflictRosterStore{},
			wantErr: ports.ErrVersionConflict,
		},
	}

	for _, tc := range failureTestCases {
		handler := roster.NewAddPlayerHandler(tc.store, testkit.NewStubLeagueLock())
		cmd := roster.NewAddPlayerCommand(testkit.TeamA(), 1)

		err := handler.Handle(cmd)

		assert.ErrorIs(t, err, tc.wantErr)
	}
}

func generateRosterHistory(id domain.TeamID, players int) []domain.RosterEvent {
	history := make([]domain.RosterEvent, players)
	for i := range players {
		history[i] = domain.AddedPlayerToRoster{
			TeamID:      id,
			PlayerID:    domain.PlayerID(i + 1),
			EffectiveAt: testkit.TodayLock(),
		}
	}

	return history
}
