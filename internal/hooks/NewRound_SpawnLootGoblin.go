package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
)

func SpawnLootGoblin(e events.Event) bool {

	evt := e.(events.NewRound)

	//
	// Load the loot goblin room (which should also spawn it), if it's time
	//
	if evt.Config.LootGoblinRoom != 0 {
		if evt.RoundNumber%uint64(evt.Config.LootGoblinRoundCount) == 0 {
			if room := rooms.LoadRoom(int(evt.Config.LootGoblinRoom)); room != nil { // loot goblin room
				mudlog.Info(`Loot Goblin Spawn`, `roundNumber`, evt.RoundNumber)
				room.Prepare(false) // Make sure the loot goblin spawns.
			}
		}
	}

	return true
}
