// Round ticks for players
package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
)

//
// Cleans up zombie users
//

func CleanupZombies_Listener(e events.Event) bool {

	evt := e.(events.NewTurn)

	expTurns := uint64(evt.Config.SecondsToTurns(int(evt.Config.ZombieSeconds)))

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
