package roster

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/ports"
)

type AddPlayerHandler struct {
	Store ports.RosterStore
	Lock  ports.LeagueLock
}

func (h AddPlayerHandler) Handle(cmd AddPlayerCommand) error {
	committed, version, err := h.Store.Load(cmd.TeamID)
	if err != nil {
		return err
	}

	stream := NewRosterStream(cmd.TeamID, committed)
	view := stream.ProjectThrough(h.Lock.NextLock())

	events, err := view.DecideAddPlayer(cmd.PlayerID)
	if err != nil {
		return err
	}

	_, err = h.Store.Append(cmd.TeamID, events, version)
	if err != nil {
		return err
	}

	return nil
}

func NewAddPlayerHandler(store ports.RosterStore, lock ports.LeagueLock) AddPlayerHandler {
	return AddPlayerHandler{
		Store: store,
		Lock:  lock,
	}
}

type AddPlayerCommand struct {
	TeamID   domain.TeamID
	PlayerID domain.PlayerID
}

func NewAddPlayerCommand(teamID domain.TeamID, playerID domain.PlayerID) AddPlayerCommand {
	return AddPlayerCommand{
		TeamID:   teamID,
		PlayerID: playerID,
	}
}
