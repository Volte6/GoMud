package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Show(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		response.SendUserMessage(userId, "Show what? To whom?", true)
		response.Handled = true
		return response, nil
	}

	var showItem items.Item = items.Item{}
	var found bool = false

	var targetName string = args[len(args)-1]
	args = args[:len(args)-1]
	var objectName string = strings.Join(args, " ")

	// Check whether the user has an item in their inventory that matches
	showItem, found = user.Character.FindInBackpack(objectName)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to show.", objectName), true)
		response.Handled = true
		return response, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// Tell the shower
			response.SendUserMessage(userId,
				fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, showItem.Name(), targetUser.Character.Name),
				true)

			// Tell the Showee
			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows you their <ansi fg="item">%s</ansi>.`, user.Character.Name, showItem.Name()),
				true)

			response.SendUserMessage(targetUser.UserId,
				"\n"+showItem.GetLongDescription()+"\n",
				true)

			// Tell the rest of the room
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, user.Character.Name, showItem.Name(), targetUser.Character.Name),
				true,
				targetUser.UserId)

		} else {
			response.SendUserMessage(userId, "Something went wrong.", true)
		}

		response.Handled = true
		return response, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				response.SendUserMessage(userId,
					fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, showItem.Name(), targetMob.Character.Name),
					true)

				// Do trigger of onShow
				if res, err := scripting.TryMobScriptEvent(`onShow`, targetMob.InstanceId, userId, `user`, map[string]any{`gold`: 0, `item`: showItem}, cmdQueue); err == nil {
					response.AbsorbMessages(res)
				}

				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, showItem.Name(), targetMob.Character.Name),
					true)

			} else {
				response.SendUserMessage(userId, "Something went wrong.", true)
			}

		}

		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, "Who???", true)

	response.Handled = true
	return response, nil
}
