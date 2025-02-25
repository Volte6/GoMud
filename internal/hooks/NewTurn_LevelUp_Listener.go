package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

//
// Check all users for a level up
//

func LevelUp_Listener(e events.Event) bool {

	evt := e.(events.NewTurn)

	// Only checks every 1 second or so
	if evt.TurnNumber%uint64(evt.Config.TurnsPerSecond()) != 0 {
		return true
	}

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		if newLevel, statsDelta := user.Character.LevelUp(); newLevel {

			livesBefore := user.Character.ExtraLives

			if evt.Config.PermaDeath && evt.Config.LivesOnLevelUp > 0 {
				user.Character.ExtraLives += int(evt.Config.LivesOnLevelUp)
				if user.Character.ExtraLives > int(evt.Config.LivesMax) {
					user.Character.ExtraLives = int(evt.Config.LivesMax)
				}
			}

			user.EventLog.Add(`xp`, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now <ansi fg="magenta-bold">level %d</ansi>!`, user.Character.Name, user.Character.Level))

			levelUpData := map[string]interface{}{
				"level":          user.Character.Level,
				"statsDelta":     statsDelta,
				"trainingPoints": 1,
				"statPoints":     1,
				"livesUp":        user.Character.ExtraLives - livesBefore,
			}
			levelUpStr, _ := templates.Process("character/levelup", levelUpData)

			user.SendText(levelUpStr)

			events.AddToQueue(events.Broadcast{
				Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> <ansi fg="yellow">has leveled up to level %d!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, user.Character.Level, term.CRLFStr),
			})

			if user.Character.Level >= 5 {
				for _, mobInstanceId := range user.Character.CharmedMobs {
					if mob := mobs.GetInstance(mobInstanceId); mob != nil {

						if mob.MobId == 38 {
							mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
							mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
							mob.Command(`suicide vanish`, 10)
						}
					}
				}
			}

			user.PlaySound(`levelup`, `other`)

			users.SaveUser(*user)

			continue
		}

	}

	return true
}
