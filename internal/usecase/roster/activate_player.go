package roster

import (
	"github.com/spcameron/dugout/internal/domain"
	"github.com/spcameron/dugout/internal/ports"
)

type ActivatePlayerHandler struct {
	Store ports.RosterStore
	Lock  ports.LeagueLock
}

func (h ActivatePlayerHandler) Handle(cmd ActivatePlayerCommand) error {
	committed, version, err := h.Store.Load(cmd.TeamID)
	if err != nil {
		return err
	}

	stream := NewRosterStream(cmd.TeamID, committed)
	view := stream.ProjectThrough(h.Lock.NextLock())

	events, err := view.DecideActivatePlayer(cmd.PlayerID, cmd.PlayerRole)
	if err != nil {
		return err
	}

	_, err = h.Store.Append(cmd.TeamID, events, version)
	if err != nil {
		return err
	}

	return nil
}

func NewActivatePlayerHandler(store ports.RosterStore, lock ports.LeagueLock) ActivatePlayerHandler {
	return ActivatePlayerHandler{
		Store: store,
		Lock:  lock,
	}
}

type ActivatePlayerCommand struct {
	TeamID     domain.TeamID
	PlayerID   domain.PlayerID
	PlayerRole domain.PlayerRole
}

func NewActivatePlayerCommand(teamID domain.TeamID, playerID domain.PlayerID, role domain.PlayerRole) ActivatePlayerCommand {
	return ActivatePlayerCommand{
		TeamID:     teamID,
		PlayerID:   playerID,
		PlayerRole: role,
	}
}
