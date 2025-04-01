package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

//
// Watches the rounds go by
// Applies autohealing where appropriate
//

func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	// Every 3 rounds. Else, pass it along.
	if evt.RoundNumber%3 != 0 {
		return events.Continue
	}

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		// Only heal if not in combat
		if user.Character.Aggro != nil {
			continue
		}

		healthStart := user.Character.Health

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

		// If it has changed, send an update
		if user.Character.Health-healthStart != 0 {

			// Trigger a redraw, but only if the users prompt has changed.
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId, OnlyIfChanged: true}, 100)

			events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})

		}

	}

	return events.Continue
}
