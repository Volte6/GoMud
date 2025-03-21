package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
)

//
// Respawn any enemies that have been missing for too long
//

func HandleRespawns(e events.Event) events.EventReturn {

	for _, roomId := range rooms.GetRoomsWithPlayers() {

		// Get rooom
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}

		room.Prepare(false)
	}

	return events.Continue
}
