// Round ticks for players
package hooks

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
)

//
// Handle mobs that are bored
//

func InactivePlayers(e events.Event) bool {

	evt := e.(events.NewRound)

	c := evt.Config.(configs.Config)

	maxIdleRounds := c.SecondsToRounds(int(c.MaxIdleSeconds))
	if maxIdleRounds == 0 {
		return true
	}

	if evt.RoundNumber < uint64(maxIdleRounds) {
		return true
	}

	kickMods := bool(configs.GetConfig().TimeoutMods)

	cutoffRound := evt.RoundNumber - uint64(maxIdleRounds)

	for _, user := range users.GetAllActiveUsers() {

		if !kickMods && user.Permission == users.PermissionAdmin || user.Permission == users.PermissionMod {
			continue
		}

		li := user.GetLastInputRound()

		if li == 0 {
			continue
		}

		if li-cutoffRound == 5 {
			user.SendText(`<ansi fg="203">WARNING:</ansi> <ansi fg="208">You are about to be kicked for inactivity!</ansi>`)
		}

		if li < cutoffRound {
			events.AddToQueue(events.System{
				Command: `kick`,
				Data:    user.UserId,
			})
		}

	}

	return true

}
