package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

//
// Respawn any enemies that have been missing for too long
//

func HandleRespawns(e events.Event) events.ListenerReturn {

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
