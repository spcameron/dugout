package roster_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/testsupport/assert"
	"github.com/spcameron/dugout/internal/testsupport/require"
	"github.com/spcameron/dugout/internal/testsupport/testkit"
	"github.com/spcameron/dugout/internal/usecase/roster"
)

func TestHandleAddPlayer(t *testing.T) {
	leagueLock := testkit.NewStubLeagueLock()
	store := testkit.NewFakeRosterStore()
	spy := testkit.NewSpyRosterStore(store)

	handler := roster.NewAddPlayerHandler(spy, leagueLock)

	cmd := roster.AddPlayerCommand{
		TeamID:   testkit.TeamA(),
		PlayerID: 1,
	}

	err := handler.Handle(cmd)

	assert.Nil(t, err)
	require.Equal(t, len(spy.AppendCalls), 1)
	require.Equal(t, spy.AppendCalls[0].TeamID, testkit.TeamA())
}
