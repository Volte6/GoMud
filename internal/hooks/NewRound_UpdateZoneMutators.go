package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

//
// Check all zones and update their mutators.
//

func UpdateZoneMutators(e events.Event) events.ListenerReturn {
	evt := e.(events.NewRound)

	// Update all zone based mutators once a round
	_, mutZoneRoomIds := rooms.GetZonesWithMutators()
	for _, rid := range mutZoneRoomIds {
		if r := rooms.LoadRoom(rid); r != nil {
			r.ZoneConfig.Mutators.Update(evt.RoundNumber)
		}
	}

	return events.Continue
}
