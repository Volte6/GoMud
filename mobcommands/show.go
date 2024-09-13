package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Show(rest string, mobId int) (util.MessageQueue, error) {
	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		response.Handled = true
		return response, nil
	}

	var showItem items.Item = items.Item{}
	var found bool = false

	var targetName string = args[len(args)-1]
	args = args[:len(args)-1]
	var objectName string = strings.Join(args, " ")

	// Check whether the user has an item in their inventory that matches
	showItem, found = mob.Character.FindInBackpack(objectName)

	if !found {
		response.Handled = true
		return response, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// Tell the Showee
			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, showItem.DisplayName()),
				true)

			response.SendUserMessage(targetUser.UserId,
				"\n"+showItem.GetLongDescription()+"\n",
				true)

			// Tell the rest of the room
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, mob.Character.Name, showItem.DisplayName(), targetUser.Character.Name),
				true,
				targetUser.UserId)

		}

		response.Handled = true
		return response, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, showItem.DisplayName(), targetMob.Character.Name),
					true)

			}

		}

		response.Handled = true
		return response, nil
	}

	response.Handled = true
	return response, nil
}
