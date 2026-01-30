package roster_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/testsupport/testkit"
)

func TestHandleAddPlayer(t *testing.T) {
	store := testkit.NewFakeRosterStore()

	teamID := testkit.TeamA()
	committed, version, err := store.Load(teamID)

	handler := roster.AddPlayerHandler{
		Store: store,
	}

	cmd := roster.AddPlayerCommand{}

	err := handler.Handle(cmd)
}
