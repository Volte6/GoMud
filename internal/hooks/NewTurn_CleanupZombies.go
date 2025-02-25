// Round ticks for players
package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
)

//
// Cleans up zombie users
// Zombie users are users who have disconnected but their user/character is still in game.
//

func CleanupZombies(e events.Event) bool {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		slog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return false
	}

	c := evt.Config.(configs.Config)

	expTurns := uint64(c.SecondsToTurns(int(c.ZombieSeconds)))

	if expTurns < evt.TurnNumber {

		expZombies := users.GetExpiredZombies(evt.TurnNumber - expTurns)
		if len(expZombies) > 0 {
			slog.Info("Expired Zombies", "count", len(expZombies))
			connIds := users.GetConnectionIds(expZombies)

			for _, userId := range expZombies {
				events.AddToQueue(events.System{Command: `leaveworld`, Data: userId})
			}
			for _, connId := range connIds {
				if err := users.LogOutUserByConnectionId(connId); err != nil {
					slog.Error("Log Out Error", "connectionId", connId, "error", err)
				}
			}

		}
	}

	return true
}
