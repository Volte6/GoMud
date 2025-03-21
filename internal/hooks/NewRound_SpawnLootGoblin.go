package hooks

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
)

func SpawnLootGoblin(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	c := configs.GetLootGoblinConfig()

	//
	// Load the loot goblin room (which should also spawn it), if it's time
	//
	if c.RoomId != 0 {
		if evt.RoundNumber%uint64(c.RoundCount) == 0 {
			if room := rooms.LoadRoom(int(c.RoomId)); room != nil { // loot goblin room
				mudlog.Info(`Loot Goblin Spawn`, `roundNumber`, evt.RoundNumber)
				room.Prepare(false) // Make sure the loot goblin spawns.
			}
		}
	}

	return events.Continue
}
