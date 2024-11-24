package usercommands

import (
	"errors"
	"fmt"
	"math"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Suicide(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	config := configs.GetConfig()

	if user.Character.Zone == `Shadow Realm` {
		user.SendText(`You're already dead!`)
		return true, errors.New(`already dead`)
	}

	if user.Character.HasBuffFlag(buffs.ReviveOnDeath) {

		user.Character.Health = user.Character.HealthMax.Value

		user.SendText(`You are revived in a shower of magical sparks!`)
		room.SendText(`<ansi fg="username">`+user.Character.Name+`</ansi> is suddenly revived in a shower of sparks!`, user.UserId)

		user.Character.CancelBuffsWithFlag(buffs.ReviveOnDeath)

		return true, nil
	}

	events.AddToQueue(events.Broadcast{
		Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>`, user.Character.Name),
	})

	// If permadeath is enabled, do some extra bookkeeping
	if config.PermaDeath {

		if user.Character.ExtraLives > 0 {

			user.Character.ExtraLives--

		} else {

			user.EventLog.Add(`death`, fmt.Sprintf(`You (<ansi fg="username">%s</ansi>) has <ansi fg="red-bold">PERMA-DIED</ansi>`, user.Character.Name))

			// Perma-died!!!
			textOut, _ := templates.Process("character/permadeath", nil)
			user.SendText(colorpatterns.ApplyColorPattern(textOut, `red`))

			// Unequip everything
			for _, itm := range user.Character.GetAllWornItems() {
				Remove(itm.Name(), user, room)
			}
			// drop all items / gold
			Drop("all", user, room)

			rooms.MoveToRoom(user.UserId, -1)

			user.Character = characters.New()

			return true, nil
		}

	}

	user.EventLog.Add(`death`, fmt.Sprintf(`You (<ansi fg="username">%s</ansi>) has <ansi fg="red-bold">DIED</ansi>`, user.Character.Name))

	if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllWornItems() {
			if util.Rand(100) < chanceInt {

				Remove(itm.Name(), user, room)

				Drop(itm.Name(), user, room)

			}
		}
	}

	if user.Character.Gold > 0 {
		user.EventLog.Add(`death`, fmt.Sprintf(`You dropped <ansi fg="gold">%d gold</ansi> on death`, user.Character.Gold))
		Drop(fmt.Sprintf(`%d gold`, user.Character.Gold), user, room)
	}

	if config.OnDeathAlwaysDropBackpack {
		Drop("all", user, room)

		user.EventLog.Add(`death`, `You dropped <ansi fg="alert-3">everthing in your backpack</ansi> on death`)

	} else if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllBackpackItems() {
			if util.Rand(100) < chanceInt {
				Drop(itm.Name(), user, room)
				user.EventLog.Add(`death`, fmt.Sprintf(`You dropped your <ansi fg="itemname">%s</ansi> on death`, itm.Name()))
			}
		}
	}

	if user.Character.Level > 1 {

		setting, lossPct := config.GetDeathXPPenalty()
		if setting != `none` {

			if setting == `level` { // are they being brought down to the base of their current level?
				user.Character.Level--
				oldExperience := user.Character.Experience
				user.Character.Experience = user.Character.XPTNL()
				user.Character.Level++

				user.SendText(fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, oldExperience-user.Character.Experience))

				user.EventLog.Add(`death`, fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>. on death`, oldExperience-user.Character.Experience))

			} else if lossPct > 0 { // Are they losing a set %?

				loss := int(math.Floor(float64(user.Character.Experience) * lossPct))
				user.Character.Experience -= loss

				user.SendText(fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, loss))

				user.EventLog.Add(`death`, fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>. on death`, loss))
			}
		}

	}

	user.Character.CancelBuffsWithFlag(buffs.All)

	user.Character.Health = -10

	user.Character.KD.AddDeath()

	rooms.MoveToRoom(user.UserId, 75)

	return true, nil
}
