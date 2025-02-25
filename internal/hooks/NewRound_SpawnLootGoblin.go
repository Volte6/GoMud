package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
)

//
// Watches the rounds go by
// Applies autohealing where appropriate
//

func SpawnLootGoblin(e events.Event) bool {

	evt := e.(events.NewRound)

	c := evt.Config.(configs.Config)

	//
	// Load the loot goblin room (which should also spawn it), if it's time
	//
	if c.LootGoblinRoom != 0 && evt.RoundNumber%uint64(c.LootGoblinRoundCount) == 0 {
		if room := rooms.LoadRoom(int(c.LootGoblinRoom)); room != nil { // loot goblin room
			slog.Info(`Loot Goblin Spawn`, `roundNumber`, evt.RoundNumber)
			room.Prepare(false) // Make sure the loot goblin spawns.
		}
	}

	return true
}
