package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

//
// Watches the rounds go by
// Applies autohealing where appropriate
//

func AutoHeal(e events.Event) bool {

	evt := e.(events.NewRound)

	// Every 3 rounds. Else, pass it along.
	if evt.RoundNumber%3 != 0 {
		return true
	}

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		// Only heal if not in combat
		if user.Character.Aggro != nil {
			continue
		}

		if user.Character.Health < 1 {
			if user.Character.RoomId != 75 {

				if user.Character.Health <= -10 {

					user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

				} else {
					user.Character.Health--
					user.SendText(`<ansi fg="red">you are bleeding out!</ansi>`)
					if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
						room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`, user.Character.Name), user.UserId)
					}
				}

			}
		} else {

			if user.Character.Health > 0 {
				user.Character.Heal(
					user.Character.HealthPerRound(),
					user.Character.ManaPerRound(),
				)
			}
		}

		newcmdprompt := user.GetCommandPrompt(true)
		oldcmdprompt := user.GetTempData(`cmdprompt`)

		// If the prompt hasn't changed, skip redrawing
		if oldcmdprompt != nil && oldcmdprompt.(string) == newcmdprompt {
			continue
		}

		// save the new prompt for next time we want to check
		user.SetTempData(`cmdprompt`, newcmdprompt)

		connections.SendTo([]byte(templates.AnsiParse(newcmdprompt)), user.ConnectionId())

		//
		// Send GMCP status update
		//
		if connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Char`) {

			realXPNow, realXPTNL := user.Character.XPTNLActual()

			events.AddToQueue(events.GMCPOut{
				UserId: user.UserId,
				Payload: fmt.Sprintf(`Char.Vitals { "hp": "%d", "maxhp": "%d", "mp": "%d", "maxmp": "%d", "xp": "%d", "xptnl": "%d", "energy": "%d", "maxenergy": "%d" }`,
					user.Character.Health, user.Character.HealthMax.Value,
					user.Character.Mana, user.Character.ManaMax.Value,
					realXPNow, realXPTNL,
					user.Character.ActionPoints, user.Character.ActionPointsMax.Value,
				),
			})
		}

	}

	return true
}
