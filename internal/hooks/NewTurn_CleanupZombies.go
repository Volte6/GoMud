// Round ticks for players
package hooks

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/users"
)

//
// Cleans up zombie users
// Zombie users are users who have disconnected but their user/character is still in game.
//

func CleanupZombies(e events.Event) events.EventReturn {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return events.Cancel
	}

	et := configs.GetTimingConfig()
	gp := configs.GetNetworkConfig()

	expTurns := uint64(et.SecondsToTurns(int(gp.ZombieSeconds)))

	if expTurns < evt.TurnNumber {

		expZombies := users.GetExpiredZombies(evt.TurnNumber - expTurns)

		if len(expZombies) > 0 {

			mudlog.Info("Expired Zombies", "count", len(expZombies))

			for _, userId := range expZombies {
				events.AddToQueue(events.System{Command: `leaveworld`, Data: userId})
			}

		}
	}

	return events.Continue
}
