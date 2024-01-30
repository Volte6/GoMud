package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Spawn(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.spawn", nil)
		response.SendUserMessage(userId, infoOutput, false)
		response.Handled = true
		return response, nil
	}

	spawnType := args[0]
	args = args[1:]

	if len(args) >= 1 {

		if spawnType == "item" {

			itemId := items.FindItemByName(strings.Join(args, ` `))

			if itemId < 1 {
				itemId, _ = strconv.Atoi(args[0])
			}

			if itemId != 0 {

				itm := items.New(itemId)
				if itm.ItemId > 0 {
					room.AddItem(itm, false)

					response.SendUserMessage(userId,
						fmt.Sprintf(`You wave your hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, itm.Name()),
						true)
					response.SendRoomMessage(user.Character.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, user.Character.Name, itm.Name()),
						true)

					response.Handled = true
					return response, nil
				}

			}
		}

		if spawnType == "gold" {

			goldAmt, _ := strconv.Atoi(args[0])
			if goldAmt < 1 {
				goldAmt = 1
			}

			room.Gold += goldAmt

			response.SendUserMessage(userId,
				fmt.Sprintf(`You wave your hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, goldAmt),
				true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, user.Character.Name, goldAmt),
				true)

			response.Handled = true
			return response, nil
		}

		if spawnType == "mob" {

			var mobId mobs.MobId

			id, err := strconv.Atoi(args[0])

			if err != nil {
				mobId = mobs.MobIdByName(strings.Join(args, ` `))
			} else {
				mobId = mobs.MobId(id)
			}

			if mobId > 0 {
				if mob := mobs.NewMobById(mobId, room.RoomId); mob != nil {
					room.AddMob(mob.InstanceId)

					response.SendUserMessage(userId,
						fmt.Sprintf(`You wave your hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, mob.Character.Name),
						true)
					response.SendRoomMessage(user.Character.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, user.Character.Name, mob.Character.Name),
						true)

					response.Handled = true
					return response, nil
				}
			}

		}

	}

	response.SendUserMessage(userId,
		"You wave your hands around pathetically.",
		true)
	response.SendRoomMessage(user.Character.RoomId,
		fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around pathetically.`, user.Character.Name),
		true)

	response.Handled = true
	return response, nil
}
