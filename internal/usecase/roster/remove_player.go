package roster

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/ports"
)

type RemovePlayerHandler struct {
	Store ports.RosterStore
	Lock  ports.LeagueLock
}

func (h RemovePlayerHandler) Handle(cmd RemovePlayerCommand) error {
	committed, version, err := h.Store.Load(cmd.TeamID)
	if err != nil {
		return err
	}

	stream := NewRosterStream(cmd.TeamID, committed)
	view := stream.ProjectThrough(h.Lock.NextLock())

	events, err := view.DecideRemovePlayer(cmd.PlayerID)
	if err != nil {
		return err
	}

	_, err = h.Store.Append(cmd.TeamID, events, version)
	if err != nil {
		return err
	}

	return nil
}

func NewRemovePlayerHandler(store ports.RosterStore, lock ports.LeagueLock) RemovePlayerHandler {
	return RemovePlayerHandler{
		Store: store,
		Lock:  lock,
	}
}

type RemovePlayerCommand struct {
	TeamID   domain.TeamID
	PlayerID domain.PlayerID
}

func NewRemovePlayerCommand(teamID domain.TeamID, playerID domain.PlayerID) RemovePlayerCommand {
	return RemovePlayerCommand{
		TeamID:   teamID,
		PlayerID: playerID,
	}
}
