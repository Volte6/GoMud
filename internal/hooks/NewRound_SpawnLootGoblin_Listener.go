package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
)

//
// Watches the rounds go by
// Applies autohealing where appropriate
//

func SpawnLootGoblin_Listener(e events.Event) bool {

	evt := e.(events.NewRound)

	//
	// Load the loot goblin room (which should also spawn it), if it's time
	//
	if evt.Config.LootGoblinRoom != 0 && evt.RoundNumber%uint64(evt.Config.LootGoblinRoundCount) == 0 {
		if room := rooms.LoadRoom(int(evt.Config.LootGoblinRoom)); room != nil { // loot goblin room
			slog.Info(`Loot Goblin Spawn`, `roundNumber`, evt.RoundNumber)
			room.Prepare(false) // Make sure the loot goblin spawns.
		}
	}

	return true
}
