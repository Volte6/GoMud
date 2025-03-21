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

func InactivePlayers(e events.Event) events.EventReturn {

	evt := e.(events.NewRound)

	c := configs.GetConfig()

	maxIdleRounds := c.Timing.SecondsToRounds(int(c.Network.MaxIdleSeconds))
	if maxIdleRounds == 0 {
		return events.Continue
	}

	if evt.RoundNumber < uint64(maxIdleRounds) {
		return events.Continue
	}

	kickMods := bool(c.Network.TimeoutMods)

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

	return events.Continue

}
