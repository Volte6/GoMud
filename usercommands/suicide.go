package usercommands

import (
	"errors"
	"fmt"
	"math"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Suicide(rest string, userId int) (bool, error) {

	config := configs.GetConfig()

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Zone == `Shadow Realm` {
		user.SendText(`You're already dead!`)
		return true, errors.New(`already dead`)
	}

	events.AddToQueue(events.Broadcast{
		Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>`, user.Character.Name),
	})

	if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllWornItems() {
			if util.Rand(100) < chanceInt {

				Remove(itm.Name(), userId)

				Drop(itm.Name(), userId)

			}
		}
	}

	if user.Character.Gold > 0 {
		Drop(fmt.Sprintf(`%d gold`, user.Character.Gold), userId)
	}

	if config.OnDeathAlwaysDropBackpack {
		Drop("all", userId)
	} else if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllBackpackItems() {
			if util.Rand(100) < chanceInt {
				Drop(itm.Name(), userId)
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
			} else if lossPct > 0 { // Are they losing a set %?

				loss := int(math.Floor(float64(user.Character.Experience) * lossPct))
				user.Character.Experience -= loss

				user.SendText(fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, loss))
			}
		}

	}

	user.Character.CancelBuffsWithFlag(buffs.All)

	user.Character.Health = -10

	user.Character.KD.AddDeath()

	rooms.MoveToRoom(userId, 75)

	return true, nil
}
