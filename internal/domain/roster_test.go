package domain_test

import (
	"testing"

	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/testutil/require"
)

func TestCanAddPlayer(t *testing.T) {
	t.Run("add player to empty roster", func(t *testing.T) {
		r := roster(10)
		err := domain.CanAddPlayer(r, 1)
		require.NoError(t, err)
	})

	t.Run("add player when already on roster", func(t *testing.T) {
		r := roster(10, 1)
		err := domain.CanAddPlayer(r, 1)
		require.ErrorIs(t, err, domain.ErrPlayerAlreadyOnRoster)
	})

	t.Run("add a second player to roster", func(t *testing.T) {
		r := roster(10, 1)
		err := domain.CanAddPlayer(r, 2)
		require.NoError(t, err)
	})

	t.Run("add a player to 26-man roster", func(t *testing.T) {
		ids := make([]domain.MLBPlayerID, domain.MaxRosterSize)
		for i := 1; i <= domain.MaxRosterSize; i++ {
			ids = append(ids, domain.MLBPlayerID(i))
		}

		r := roster(10, ids...)
		err := domain.CanAddPlayer(r, 999)
		require.ErrorIs(t, err, domain.ErrRosterFull)
	})
}

func roster(teamID domain.TeamID, ids ...domain.MLBPlayerID) domain.Roster {
	r := domain.Roster{
		TeamID:  teamID,
		Entries: make([]domain.RosterEntry, len(ids)),
	}

	for i, id := range ids {
		r.Entries[i] = domain.RosterEntry{
			MLBID: id,
		}
	}

	return r
}
