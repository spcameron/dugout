package roster

import "github.com/spcameron/dugout/internal/ports"

type AddPlayerHandler struct {
	Store ports.RosterStore
}

func (h AddPlayerHandler) Handle() {}
