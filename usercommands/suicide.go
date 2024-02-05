package usercommands

import (
	"errors"
	"fmt"
	"math"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Suicide(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	config := configs.GetConfig()

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Zone == `Shadow Realm` {
		response.SendUserMessage(userId, `You're already dead!`, true)
		response.Handled = true
		return response, errors.New(`already dead`)
	}

	response.SendRoomMessage(0, fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>`, user.Character.Name), true)

	if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllWornItems() {
			if util.Rand(100) < chanceInt {

				resp, _ := Remove(itm.Name(), userId, cmdQueue)
				response.AbsorbMessages(resp)

				resp, _ = Drop(itm.Name(), userId, cmdQueue)
				response.AbsorbMessages(resp)

			}
		}
	}

	user.Character.Gold = 0

	if config.OnDeathAlwaysDropBackpack {
		resp, _ := Drop("all", userId, cmdQueue)
		response.AbsorbMessages(resp)
	} else if config.OnDeathEquipmentDropChance >= 0 {
		chanceInt := int(config.OnDeathEquipmentDropChance * 100)
		for _, itm := range user.Character.GetAllBackpackItems() {
			if util.Rand(100) < chanceInt {

				resp, _ := Drop(itm.Name(), userId, cmdQueue)
				response.AbsorbMessages(resp)

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

				response.SendUserMessage(userId, fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, oldExperience-user.Character.Experience), true)
			} else if lossPct > 0 { // Are they losing a set %?

				loss := int(math.Floor(float64(user.Character.Experience) * lossPct))
				user.Character.Experience -= loss

				response.SendUserMessage(userId, fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, loss), true)
			}
		}

	}

	user.Character.CancelBuffsWithFlag(buffs.All)

	user.Character.Health = -10

	rooms.MoveToRoom(userId, 75)

	response.Handled = true
	return response, nil
}
